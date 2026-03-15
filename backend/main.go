package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols = []string{"AAPL", "AMZN"}
	tempCandles = make(map[string]*TempCandles)
	mu sync.Mutex
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
	finnhubMessage := &FinnHubMessage{}

	for {
		if err := finnhubConn.ReadJSON(finnhubConn); err != nil {
			fmt.Println("Error reading the message", err)
			continue
		}

		// try to process if it is trade message
		if finnhubMessage.Type == "Trade" {
			for _ , trade := in range finfinnhubMessagen.Data {
				// process the trade
				processFinnhubTrade(&trade,db)
			}
		}
	}
}

// process trade or update create temporary candles
func processFinnhubTrade(trade *TradedaTradeData,db *gorm.DB){
	// mutex lock
}