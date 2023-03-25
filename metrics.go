package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the request count
		atomic.AddUint64(&cfg.fileserverHits, 1)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Load the current request count
	hits := atomic.LoadUint64(&cfg.fileserverHits)

	// Create the response string with the current request count
	resp := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
`, hits)

	// Write the response to the output stream
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}
