package main

import (
	"net/http"
)

func main() {
	// Create a new http.ServeMux
	mux := http.NewServeMux()

	// add a handler for the root path
	mux.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8000", mux)

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middlewareCors(mux)

	// Create a new http.Server and use the corsMux as the handler
	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}

	// Use the server's ListenAndServe method to start the server
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
