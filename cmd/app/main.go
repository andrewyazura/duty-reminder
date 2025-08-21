package main

import (
	"io"
	"log"
	"net/http"

	"github.com/andrewyazura/duty-reminder/internal/config"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ok")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v\n", r)

		next.ServeHTTP(w, r)
	})
}

func main() {
	c := config.NewConfig()

	err := config.LoadJSONConfigFile(c, "config.json")
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	http.HandleFunc("/health", healthCheck)

	http.ListenAndServe(":1234", LoggingMiddleware(http.DefaultServeMux))
}
