package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"
	// Create a new http.ServeMux
	mux := http.NewServeMux()

	// Add a handler for the root path
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middlewareCors(mux)

	// Create a new http.Server and use the corsMux as the handler
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	// Use the server's ListenAndServe method to start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
