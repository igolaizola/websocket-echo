package wsecho

import (
	"context"
	"log"
)

// Server serves the wsecho server.
func Serve(ctx context.Context, port int) error {
	log.Printf("server listening on port %d\n", port)
	<-ctx.Done()
	return nil
}
