package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits uint64
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	apiCfg := &apiConfig{}
	// Create a new http.ServeMux
	mux := http.NewServeMux()

	// Wrap the http.FileServer handler with a custom middleware function that increments the request count
	mux.Handle("/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))))

	// Add a handler for the /healthz path
	mux.HandleFunc("/healthz", healthzHandler)

	// Add a handler for the /metrics path that returns the request count as plain text
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)

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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the request count
		atomic.AddUint64(&cfg.fileserverHits, 1)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Write the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write the status code using w.WriteHeader
	w.WriteHeader(http.StatusOK)

	// Write the body text using w.Write
	fmt.Fprintf(w, "Hits: %d", atomic.LoadUint64(&cfg.fileserverHits))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Write the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write the status code using w.WriteHeader
	w.WriteHeader(http.StatusOK)

	// Write the body text using w.Write
	w.Write([]byte("OK"))
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
