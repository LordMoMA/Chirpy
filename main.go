package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/lordmoma/chirpy/internal/database"
	"github.com/lordmoma/chirpy/internal/handlers"
	"github.com/lordmoma/chirpy/internal/middleware"
)

func main() {
	godotenv.Load()
	// // use flag package in Go to parse command line flags
	// debug := flag.Bool("debug", false, "enable debugging") // create a boolean value for the --debug flag

	// flag.Parse() // parse the command line flags
	// if *debug {  // check the value of the debug flag
	// 	fmt.Println("Debugging enabled")
	// } else {
	// 	fmt.Println("Debugging disabled")
	// }

	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	apiCfg := &handlers.ApiConfig{}

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
	apiRouter.Get("/healthz", handlers.HealthzHandler)
	apiRouter.Post("/chirps", handlers.CreateChirpsHandler(db))
	apiRouter.Get("/chirps", handlers.GetChirpsHandler(db))
	apiRouter.Get("/chirps/{id}", handlers.GetChirpIDHandler(db))

	// create users for /api namespaces
	apiRouter.Post("/users", handlers.CreateUserHandler(db))
	apiRouter.Post("/login", handlers.LoginHandler(db))

	// create a new router for the admin
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", handlers.MetricsHandler(apiCfg))

	// Mount the apiRouter at /api in the main router
	r := chi.NewRouter()
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	// Serve static files from the root directory and add the middleware to track metrics
	r.Mount("/", middleware.MiddlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)), apiCfg))

	// Wrap the mux in a custom middleware function that adds CORS headers to the response
	corsMux := middleware.MiddlewareCors(r)

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
