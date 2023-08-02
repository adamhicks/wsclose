package wsclose

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"sync"
	"time"
)

func SyncWait(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	return ch
}

type WebSocketStreams struct {
	wsStreams sync.WaitGroup
}

func (ws *WebSocketStreams) WS(work func(ctx context.Context, conn *websocket.Conn) error) httprouter.Handle {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: time.Second * 10,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("hey request!")
		// Add the goroutine before calling http.Hijack, this way the http Shutdown
		// will wait until we've called Add for all potential websocket streams
		// before it returns. A call to wsStreams.Wait will therefore wait for all
		// active websockets to close.
		ws.wsStreams.Add(1)
		defer ws.wsStreams.Done()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		err = work(r.Context(), conn)
		if err != nil {
			log.Println(err)
		}
	}
}

func (ws *WebSocketStreams) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-SyncWait(&ws.wsStreams):
		return nil
	}
}

func ShutdownAll(ctx context.Context, srv *http.Server, streams *WebSocketStreams) error {
	log.Println("shutting down")
	err := srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	log.Println("waiting for streams")
	return streams.Shutdown(ctx)
}
