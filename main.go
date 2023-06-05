package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
	upgrader  = websocket.Upgrader{}
)

// Message estructura para los mensajes del chat
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		clients[conn] = true

		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients, conn)
				break
			}
			broadcast <- msg
		}
	})

	go handleMessages()

	router.Run(":8080")
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
