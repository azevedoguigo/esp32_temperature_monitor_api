package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := gin.Default()

	router.GET("/ws", func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			return
		}

		defer conn.Close()

		for {
			_, temperature, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error to read temperature!")
				break
			}

			fmt.Println("Actual temperature: ", string(temperature))
		}
	})

	router.Run()
}
