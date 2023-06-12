package wsecho

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Server serves the wsecho server.
func Serve(ctx context.Context, addr string) error {
	log.Printf("server listening on %s\n", addr)

	// Create a new server.
	srv := &http.Server{
		Addr:    addr,
		Handler: NewServer(),
	}

	// Listen until the context is cancelled.
	go func() {
		<-ctx.Done()
		log.Println("server shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("couldn't shutdown: %v\n", err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("couldn't serve: %w", err)
	}
	return nil
}

type Server struct {
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// ServeHTTP implements http.Handler.ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Websocket connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(fmt.Errorf("couldn't upgrade: %w", err))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(fmt.Errorf("couldn't close: %w", err))
		}
	}()

	// Ping pong handlers
	conn.SetPingHandler(func(appData string) error {
		// Send pong
		log.Printf("ping: %s\n", appData)
		return conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})
	conn.SetPongHandler(func(appData string) error {
		log.Printf("pong: %s\n", appData)
		return nil
	})

	// Close handler
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("close: %d %s\n", code, text)
		cancel()
		return nil
	})

	// Echo messages
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(fmt.Errorf("couldn't read: %w", err))
			break
		}
		log.Printf("recv: %d bytes", len(message))
		if err := conn.WriteMessage(mt, message); err != nil {
			log.Println(fmt.Errorf("couldn't write: %w", err))
			break
		}
	}
}

func Ping(ctx context.Context, host string, n, size int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Connect to the websocket server.
	conn, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err != nil {
		return fmt.Errorf("couldn't dial: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(fmt.Errorf("couldn't close: %w", err))
		}
	}()

	// Ping pong handlers
	conn.SetPingHandler(func(appData string) error {
		// Send pong
		log.Printf("ping: %s\n", appData)
		return conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})
	conn.SetPongHandler(func(appData string) error {
		log.Printf("pong: %s\n", appData)
		return nil
	})

	// Close handler
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("close: %d %s\n", code, text)
		cancel()
		return nil
	})

	// Send data
	var elapsed []time.Duration
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		start := time.Now()
		if err := conn.WriteMessage(websocket.BinaryMessage, make([]byte, size)); err != nil {
			return fmt.Errorf("couldn't write: %w", err)
		}
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(fmt.Errorf("couldn't read: %w", err))
			break
		}
		elapsed := time.Since(start)
		log.Printf("sent %d bytes in %s\n", size, elapsed)
	}

	// Print average
	if len(elapsed) > 0 {
		var sum time.Duration
		for _, d := range elapsed {
			sum += d
		}
		log.Println("average:")
		log.Printf("sent %d bytes in %s\n", size*len(elapsed), sum/time.Duration(len(elapsed)))
	}
	return nil
}
