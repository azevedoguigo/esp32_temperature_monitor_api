package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TemperatureStore struct {
	mutex           sync.Mutex
	lastTemperature string
	clients         map[*websocket.Conn]bool
}

var temperatureStore = TemperatureStore{
	clients: make(map[*websocket.Conn]bool),
}

func main() {
	router := gin.Default()

	router.GET("/send-temperature", func(ctx *gin.Context) {
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

			temperatureStore.mutex.Lock()
			temperatureStore.lastTemperature = string(temperature)
			temperatureStore.mutex.Unlock()

			fmt.Println("Actual temperature: ", string(temperature))

			temperatureStore.temperatureBroadcast(string(temperature))
		}

		temperatureStore.mutex.Lock()
		delete(temperatureStore.clients, conn)
		temperatureStore.mutex.Unlock()
	})

	router.GET("/get-temperature", func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			return
		}

		temperatureStore.mutex.Lock()
		temperatureStore.clients[conn] = true
		temperatureStore.mutex.Unlock()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		temperatureStore.mutex.Lock()
		delete(temperatureStore.clients, conn)
		temperatureStore.mutex.Unlock()
	})

	router.Run()
}

func (store *TemperatureStore) temperatureBroadcast(temperature string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	for conn := range store.clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(temperature))

		if err != nil {
			fmt.Println("Error broadcasting temperature: ", err)
			conn.Close()
			delete(store.clients, conn)
		}
	}
}
