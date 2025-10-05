package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devanshu06/go-hotenv/hotenv"
)

func main() {
	// Optional: start watcher early (otherwise first Getenv() starts it)
	hotenv.Init("")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		// hotenv reads from /app/secrets/.env (or $SECRETS_FILE) and hot-reloads on change
		greeting := hotenv.Getenv("GREETING_TEXT", "Hello")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, greeting)
	})

	port := hotenv.Getenv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
