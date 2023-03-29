package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/lordmoma/chirpy/internal/database"
	"github.com/lordmoma/chirpy/internal/handlers"
	"github.com/lordmoma/chirpy/internal/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	apiCfg := &handlers.ApiConfig{
		FileserverHits: 0,
		JwtSecret:      jwtSecret,
	}
	fmt.Printf("JWT_SECRET: %s\n", apiCfg.JwtSecret)
	// use flag package in Go to parse command line flags
	debug := flag.Bool("debug", false, "enable debugging") // create a boolean value for the --debug flag

	flag.Parse() // parse the command line flags
	if *debug {  // check the value of the debug flag
		fmt.Println("Debugging enabled")
	} else {
		fmt.Println("Debugging disabled")
	}

	const filepathRoot = "."
	const port = "8080"

	// Create a new apiConfig struct to hold the request count
	// apiCfg := &handlers.ApiConfig{}

	// Create a new Database
	db, err := database.NewDB("database.json")
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("Failed to open database file")
	}
	defer os.Remove("database.json")

	// Create a new router for the /api namespace
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlers.HealthzHandler)
	apiRouter.Post("/chirps", handlers.CreateChirpsHandler(db))
	apiRouter.Get("/chirps", handlers.GetChirpsHandler(db))
	apiRouter.Get("/chirps/{id}", handlers.GetChirpIDHandler(db))

	// create users for /api namespaces
	apiRouter.Post("/users", handlers.CreateUserHandler(db))
	apiRouter.Put("/users", handlers.CreateUserHandler(db))
	apiRouter.Post("/login", handlers.LoginHandler(db, apiCfg))

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
		log.Println("ListenAndServe:", err)
	}

	// Set up an operating system signal handler to capture the Ctrl+C signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Wait for the signal
	<-signalChan

	// Shutdown the server gracefully
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}
}
