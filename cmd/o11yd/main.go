package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/example/workshop-iidp-o11y/internal/server"
)

func main() {
	addr := flag.String("listen", ":8080", "HTTP listen address")
	readTimeout := flag.Duration("read-timeout", 15*time.Second, "HTTP read timeout")
	writeTimeout := flag.Duration("write-timeout", 30*time.Second, "HTTP write timeout")
	idleTimeout := flag.Duration("idle-timeout", 60*time.Second, "HTTP idle timeout")
	flag.Parse()

	srv := &http.Server{
		Addr:         *addr,
		Handler:      server.NewHTTPHandler(),
		ReadTimeout:  *readTimeout,
		WriteTimeout: *writeTimeout,
		IdleTimeout:  *idleTimeout,
	}

	log.Printf("o11yd listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
