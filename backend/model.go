package main

import "time"

// Candle struct represents the signle OHLC(open high low close) candle
type Candle struct {
	Symbol     string    `json:"symbol"`
	Open       float64   `json:"open"`
	Close      float64   `json:"close"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Timestamps time.Time `json:"timestamps"`
}

type TempCandles struct {
	Symbol     string
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	LowPrice   float64
	HighPrice  float64
	Volume     float64
}

type FinnHubMessage struct {
	Data []TradeData `json:"data"`
	Type string      `json:"type"` // ping | trade
}

type TradeData struct {
	Close     []string `json:"c"`
	Price     float64  `json:"p"`
	Symbol    string   `json:"s"`
	Timestamp int64    `json:"t"`
	Volume    int      `json:"v"`
}

func (tc *TempCandles) toCandle() *Candle {
	return &Candle{
		Symbol:     tc.Symbol,
		Open:       tc.OpenPrice,
		Close:      tc.ClosePrice,
		High:       tc.HighPrice,
		Low:        tc.LowPrice,
		Timestamps: tc.CloseTime,
	}
}

type BroadcastMessage struct {
	UpdateType UpdateType `json:"updateType"`
	Candle     *Candle    `json:"candle"`
}

type UpdateType string

const (
	Live   UpdateType = "live"
	Closed UpdateType = "closed"
)
