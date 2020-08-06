package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	BaseURL       *url.URL
	conn          *websocket.Conn
	Logger        *log.Logger
	authenticated bool
	APIKey        string
	Secret        []byte
}

func NewWsClient(apiKey string, secret []byte) *WsClient {
	baseURL, _ := url.Parse("wss://ftx.us/ws/")
	c := &WsClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Secret:  secret,
	}
	// header := c.Authenticate()
	conn, _, err := websocket.DefaultDialer.Dial(baseURL.String(), nil)

	if err != nil {
		log.Println("Error establishing web socket connection")
		log.Fatal(err)
	}
	log.Println("Websocket connection established")
	c.conn = conn
	return c
}

func (c *WsClient) Authenticate() http.Header {
	ts := strconv.FormatInt(time.Now().UTC().Unix()*1000, 10)
	signaturePayload := ts
	mac := hmac.New(sha256.New, c.Secret)
	mac.Write([]byte(signaturePayload))
	signature := hex.EncodeToString(mac.Sum(nil))
	header := http.Header{}
	header.Set("sign", signature)
	header.Set("key", c.APIKey)
	header.Set("time", ts)

	return header
}

func (c *WsClient) Ping() {
	type msg struct {
		Op string `json:"op"`
	}

	err := c.conn.WriteJSON(msg{
		Op: "ping",
	})

	if err != nil {
		log.Println("Error reading response: ", err)
	}

	log.Println("all good")

}
