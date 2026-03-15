package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols           = []string{"AAPL", "AMZN"}
	tempCandles       = make(map[string]*TempCandles)
	mu                sync.Mutex
	broadcaster       = make(chan *BroadcastMessage)
	clientConnections = make(map[*websocket.Conn]bool)
)

func WSHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to websocket:", err)
		return
	}

	mu.Lock()
	clientConnections[conn] = true
	mu.Unlock()

	defer conn.Close()
	defer func() {
		mu.Lock()
		delete(clientConnections, conn)
		mu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message from client:", err)
			break
		}
	}
}

func main() {
	// 1. env configuration
	env := EnvConfig()
	// 2. database connection
	dbConn := DBConnection(env)

	// 3. connect to finhub websockets
	finehubConn := connectToFinHub(env)
	defer finehubConn.Close()
	// 4. handle incoming messages from finhub
	go handleFinnhubIncomingMessages(finehubConn, dbConn)
	// 5. broadcast all the clients connected
	go broadCastUpdates()
	// --- Endpoints ---

	// connect to websocket
	http.HandleFunc("/ws", WSHandler)
	// fetch all candles for all of symbols
	// fetch all candles for specific symbol

	// serve the endpoint
}

// connect to finhub
func connectToFinHub(env *Env) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://ws.finnhub.io?token=%s", env.API_KEY), nil)
	if err != nil {
		panic(err)
	}

	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		ws.WriteMessage(websocket.TextMessage, msg)
	}

	return ws
}

// handle incoming messages from finhub
func handleFinnhubIncomingMessages(finnhubConn *websocket.Conn, dbConn *gorm.DB) {
	for {
		finnhubMessage := &FinnHubMessage{}
		if err := finnhubConn.ReadJSON(finnhubMessage); err != nil {
			fmt.Println("Error reading the message", err)
			continue
		}

		// try to process if it is trade message
		if finnhubMessage.Type == "Trade" {
			for _, trade := range finnhubMessage.Data {
				// process the trade
				processFinnhubTrade(&trade, dbConn)
			}
		}
	}
}

// process trade or update create temporary candles
func processFinnhubTrade(trade *TradeData, db *gorm.DB) {
	mu.Lock()
	defer mu.Unlock()

	symbol := trade.Symbol
	price := trade.Price
	volume := float64(trade.Volume)
	timestamp := trade.Timestamp // Unix ms
	tradeTime := time.UnixMilli(timestamp)

	tempCandle, exists := tempCandles[symbol]

	if !exists {
		candle := tempCandle.toCandle()
		// save candle to db
		if err := db.Create(candle).Error; err != nil {
			fmt.Println("Error saving candle to database", err)
		}

		broadcaster <- &BroadcastMessage{UpdateType: Closed, Candle: candle}

		// initialize new temp candle
		tempCandles[symbol] = &TempCandles{
			Symbol:     symbol,
			OpenTime:   tradeTime,
			CloseTime:  tradeTime,
			OpenPrice:  price,
			ClosePrice: price,
			LowPrice:   price,
			HighPrice:  price,
			Volume:     volume,
		}
		return
	}

	// Update existing temp candle
	if price < tempCandle.LowPrice {
		tempCandle.LowPrice = price
	}
	if price > tempCandle.HighPrice {
		tempCandle.HighPrice = price
	}
	tempCandle.ClosePrice = price
	tempCandle.CloseTime = tradeTime
	tempCandle.Volume += volume

	// store the temp candle for the symbol
	tempCandles[symbol] = tempCandle

	// write the broadcast message to the channel
	broadcaster <- &BroadcastMessage{UpdateType: Live, Candle: tempCandle.toCandle()}
}

// send update every one sec until the candle is closed
func broadCastUpdates() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message := <-broadcaster:
			if message.UpdateType == Closed {
				fmt.Printf("Candle closed for %s at %s\n", message.Candle.Symbol, message.Candle.Timestamps.Format(time.RFC3339))
			} else {
				fmt.Printf("Broadcasting live update for %s at %s\n", message.Candle.Symbol, message.Candle.Timestamps.Format(time.RFC3339))
			}
			// Broadcast to all clients
			mu.Lock()
			for client := range clientConnections {
				err := client.WriteJSON(message)
				if err != nil {
					fmt.Printf("Error sending update to client: %v\n", err)
				}
			}
			mu.Unlock()
		case <-ticker.C:
			// Optionally send periodic updates or perform cleanup
			// Example: send latest candle snapshot to all clients
			mu.Lock()
			for _, tempCandle := range tempCandles {
				msg := &BroadcastMessage{UpdateType: Live, Candle: tempCandle.toCandle()}
				for client := range clientConnections {
					err := client.WriteJSON(msg)
					if err != nil {
						fmt.Printf("Error sending periodic update to client: %v\n", err)
					}
				}
			}
			mu.Unlock()
		}
	}
}

