package wsclose

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func Work(ctx context.Context, ws *websocket.Conn) error {
	defer func() {
		deadline := time.Now().Add(time.Second)
		_ = ws.WriteControl(websocket.CloseMessage, nil, deadline)
		ws.Close()
	}()

	tick := time.NewTicker(time.Second)
	var i int
	for {
		select {
		case ti := <-tick.C:
			i++
			s := fmt.Sprintf("%d: %s", i, ti)
			err := ws.WriteMessage(websocket.TextMessage, []byte(s))
			if err != nil {
				log.Println("finish on write", err)
				return err
			}
		case <-ctx.Done():
			log.Println("waiting before shutdown")
			time.Sleep(time.Second)
			return ctx.Err()
		}
	}
}
