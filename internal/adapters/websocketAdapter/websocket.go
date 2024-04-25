package websocket

import (
	adapters "Anokha-main/internal/adapters/TradingAdapter"
	longservices "Anokha-main/internal/core/Long-Trade/services"
	shortservices "Anokha-main/internal/core/Short-Trade/services"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var signalChannel = make(chan string)
var PositionChannel = make(chan string)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var (
	apiKey    = "d38ff8ff03f3cb1de0693a83fedfa520b8d9a76f6e0f6456eccaeab327618e9f"
	secretKey = "6f06f82ebbfd7a4d2a544cef65f0f558ca348f88718f5d158af2cad22f42346c"
)

type WebSockets struct{}

func (ws *WebSockets) ServeWebSocketLT(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error Connecting WebSocket")
		}
	}(conn)

	services := longservices.TradingService{}
	adapter := adapters.TradingStrategyAdapter{}
	go adapter.RunTradingStrategy(apiKey, secretKey, signalChannel)
	go services.CheckUpdateOrder()
	go services.IsPositionCompleted()

	for signal := range signalChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(signal))
		if err != nil {
			log.Println("WriteMessage:", err)
			return
		}
	}
}
func (ws *WebSockets) ServeWebSocketP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error Connecting WebSocket")
		}
	}(conn)
	adapter := adapters.TradingStrategyAdapter{}

	go adapter.RunLivePosition(PositionChannel)

	for positions := range PositionChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(positions))
		if err != nil {
			log.Println("WriteMessage:", err)
			return
		}
	}
}

func (ws *WebSockets) ServeWebSocketST(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error Connecting WebSocket")
		}
	}(conn)

	services := shortservices.TradingShortService{}
	adapter := adapters.TradingStrategyAdapter{}
	go adapter.RunShortTradingStrategy(apiKey, secretKey, signalChannel)
	go services.CheckUpdateOrder()
	go services.IsPositionCompleted()

	for signal := range signalChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(signal))
		if err != nil {
			log.Println("WriteMessage:", err)
			return
		}
	}
}
