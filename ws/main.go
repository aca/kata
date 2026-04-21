package main

import (
	"context"
	"log"
	"time"

	"github.com/coder/websocket"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:9222/devtools/browser/31b155c1-ef8b-48bf-8f75-3957778bd71b", nil)
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

	log.Println("Connected to the browser")
	time.Sleep(time.Second * 5)

	c.Close(websocket.StatusNormalClosure, "")

}
