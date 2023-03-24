package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits uint64
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	apiCfg := &apiConfig{}

	// Create a new router for the /api namespace
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzHandler)

	// create a new router for the admin
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)

	// Mount the apiRouter at /api in the main router
	r := chi.NewRouter()
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)
	r.Mount("/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))))

	// Serve static files from the root directory and add the middleware to track metrics

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middlewareCors(r)

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
