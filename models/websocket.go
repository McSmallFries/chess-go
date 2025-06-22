package models

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ChessGameSocket struct {
	Id int64 `json:"socketId" db:"SocketID"`
	Ws *websocket.Conn
}

// vars

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[int64]ChessGameSocket) // Connected clients
var broadcast = make(chan []byte)             // Broadcast channel
var mutex = &sync.Mutex{}
var wg = &sync.WaitGroup{}

func GetWaitGroup() *sync.WaitGroup {
	return wg
}

func WsHandler(w http.ResponseWriter, r *http.Request, GameId int64) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return err
	}
	defer conn.Close()

	newGame := ChessGameSocket{
		Id: GameId,
		Ws: conn,
	}
	mutex.Lock()

	clients[GameId] = newGame
	mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, GameId)
			mutex.Unlock()
			break
		}
		// put message thru broadcast channel
		broadcast <- message
	}
	return nil
}

// runs async to the incoming req handler fun above
func HandleWsMessages() {
	for {
		// Grab the next message from the broadcast channel
		message := <-broadcast

		// scan into struct, edit the global table of games
		// messageJson := string(message)

		// dont do this -->: Send the message to all connected clients
		mutex.Lock()
		for i := range clients {
			conn := clients[i].Ws
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				conn.Close()
				delete(clients, i)
			}
		}
		mutex.Unlock()
	}
}
