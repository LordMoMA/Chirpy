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

	// Add a handler for the /healthz path
	mux.HandleFunc("/healthz", healthzHandler)

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middlewareCors(mux)

	// Create a new http.Server and use the corsMux as the handler
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	// Use the server's ListenAndServe method to start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	// log.Fatal(srv.ListenAndServe())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Write the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write the status code using w.WriteHeader
	w.WriteHeader(http.StatusOK)

	// Write the body text using w.Write
	w.Write([]byte("OK"))
}
