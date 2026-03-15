package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	symbols = []string{"AAPL", "AMZN", "BINANCE:BTCUSDT", "IC MARKETS:1"}
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
