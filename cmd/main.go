package main

import (
	"context"
	"errors"
	"github.com/adamhicks/wsclose"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	runServer(ctx)
	log.Println("shutdown complete")
}

func runServer(ctx context.Context) {
	r := httprouter.New()
	ws := new(wsclose.WebSocketStreams)

	r.GET("/ws", ws.WS(wsclose.Work))

	srv := http.Server{
		Addr:        "127.0.0.1:8008",
		Handler:     r,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := wsclose.ShutdownAll(shutCtx, &srv, ws)
	if err != nil {
		log.Println(err)
	}
}
