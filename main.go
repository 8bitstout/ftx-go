package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var apiKey = os.Getenv("FTX_API_KEY")
var apiSecret = os.Getenv("FTX_API_SECRET")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Subscription struct {
	Op      string `json:"op"`
	Channel string `json:"channel"`
	Market  string `json:"market"`
}

type Response struct {
	Channel string `json:"channel"`
	Market  string `json:"market"`
	Type    string `json:"type"`
}

type OrderbookResponse struct {
	Channel string `json:"channel"`
	Market  string `json:"market"`
	Type    string `json:"type"`
	Data    struct {
		Time     float64     `json:"time"`
		Checksum int64       `json:"checksum"`
		Bids     [][]float32 `json:"bids"`
		Asks     [][]float32 `json:"asks"`
		Action   string      `json:"action"`
	} `json:"data"`
}

func (o *OrderbookResponse) getBestBidPrice() float32 {
	var price float32
	price = 0.00

	if len(o.Data.Bids) > 0 {
		price = o.Data.Bids[0][0]
	}

	return price
}

func makeSubscription(market string) *Subscription {
	m := &Subscription{}
	m.Op = "subscribe"
	m.Channel = "orderbook"
	m.Market = market

	return m
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ws := NewWsClient(apiKey, []byte(apiSecret))
	defer ws.conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, msg, err := ws.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", msg)
			obResponse := &OrderbookResponse{}
			if err := json.Unmarshal(msg, &obResponse); err != nil {
				panic(err)
			}
			if err := json.Unmarshal(msg, &obResponse); err != nil {
				panic(err)
			}
			fmt.Println(obResponse.Data)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	s := makeSubscription("ETH/USD")
	msg, _ := json.Marshal(&s)
	err := ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))

	if err != nil {
		log.Println("Error when subscribing:", err)
		return
	}

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println(t.String())
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
