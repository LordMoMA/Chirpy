package handlers

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/lordmoma/chirpy/internal/config"
)


func MetricsHandler(cfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load the current request count
		hits := atomic.LoadUint64(&cfg.FileserverHits)

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
}
