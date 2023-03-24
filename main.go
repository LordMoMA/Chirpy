package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	apiRouter.Post("/validate_chirp", apiCfg.validateHandler)

	// create a new router for the admin
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)

	// Mount the apiRouter at /api in the main router
	r := chi.NewRouter()
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	// Serve static files from the root directory and add the middleware to track metrics
	r.Mount("/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))))

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

func (cfg *apiConfig) validateHandler(w http.ResponseWriter, r *http.Request) {

	var chirp struct {
		Body string `json:"body"`
	}

	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Replace profane words with asterisks

	replaceProfane(chirp.Body)

	response := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: chirp.Body,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func replaceProfane(text string) string {
	profane := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range profane {
		text = strings.ReplaceAll(text, word, "****")
		text = strings.ReplaceAll(text, strings.ToUpper(word), "****")
	}
	return text
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (cfg *apiConfig) validateHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON body into a struct
	var chirp struct {
		Body string `json:"body"`
	}
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Check if the Chirp is too long
	if len(chirp.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Chirp is too long"})
		return
	}

	// If the Chirp is valid, respond with a success message
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"valid": true})
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
