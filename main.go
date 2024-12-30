package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewy9/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries // Holds database queries for API handlers
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	// Get PORT from environment variables
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	// Get database connection string from environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	// Initialize database connection
	connection, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connection to database: ", err)
	}

	// Create API configuration with database connection, so that it works with queries
	apiConfig := apiConfig{
		DB: database.New(connection),
	}

	// Create a new router using chi
	fmt.Println("Port:", portString)
	router := chi.NewRouter()

	// Set up CORS middleware
	// This allows cross-origin requests with specific rules
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},                   // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Allowed HTTP methods
		AllowedHeaders:   []string{"*"},                                       // Allow all headers
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Cache preflight requests for 5 minutes
	}))

	// Create v1 version of the API router
	v1Router := chi.NewRouter()

	// Define API endpoints
	v1Router.Get("/healthz", handlerReadiness)          // Health check endpoint
	v1Router.Get("/err", handlerErr)                    // Error test endpoint
	v1Router.Post("/users", apiConfig.hanlerCreateUser) // Create user endpoint

	// Mount v1 router to main router under /v1 path
	router.Mount("/v1", v1Router)

	// Configure and start HTTP server
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	// Start the server and listen for requests
	log.Printf("Server starting on port %v", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
