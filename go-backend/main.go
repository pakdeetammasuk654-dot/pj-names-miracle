package main

import (
	"log"

	"go-backend/config"
	"go-backend/delivery/http/handler"
	"go-backend/delivery/http/router"
	"go-backend/repository"
	"go-backend/usecase"
)

func main() {
	// 1. Initialize PostgreSQL Connection
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Repository
	nameRepo := repository.NewNameRepository(db)

	// 3. Initialize UseCases
	healthUseCase := usecase.NewHealthUseCase("Go-Clean-Architecture-Service")
	nameSearchUseCase := usecase.NewNameSearchUseCase(nameRepo)

	// 4. Initialize Handlers
	healthHandler := handler.NewHealthHandler(healthUseCase)
	nameHandler := handler.NewNameHandler(nameSearchUseCase)

	// 5. Setup Gin Engine & Routes
	r := router.SetupRouter(healthHandler, nameHandler)

	log.Println("Starting Gin server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
