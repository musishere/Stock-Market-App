package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols     = []string{"AAPL", "AMZN"}
	tempCandles = make(map[string]*TempCandles)
	mu          sync.Mutex
	broadcaster = make(chan *BroadcastMessage)
)

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

	// --- Endpoints ---

	// connect to websocket
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
}
