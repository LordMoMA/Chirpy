package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/lordmoma/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits uint64
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	apiCfg := &apiConfig{}

	// Create a new Database
	db, err := database.NewDB("database.json")
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("Failed to open database file")
	}

	// Create a new router for the /api namespace
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzHandler)
	apiRouter.Post("/chirps", db.CreateChirpsHandler)
	apiRouter.Get("/chirps", db.GetChirpsHandler)
	// apiRouter.Get("/chirps/", db.GetChirpsHandler)

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
