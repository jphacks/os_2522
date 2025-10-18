package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/teradatakeshishou/os_2522/backend/internal/database"
	"github.com/teradatakeshishou/os_2522/backend/internal/handler"
	"github.com/teradatakeshishou/os_2522/backend/internal/middleware"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
	"github.com/teradatakeshishou/os_2522/backend/internal/service"
)

func main() {
	// Initialize database
	dbConfig := database.GetConfigFromEnv()
	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database initialized successfully")

	// Initialize repositories
	personRepo := repository.NewPersonRepository(db)
	faceRepo := repository.NewFaceRepository(db)
	encounterRepo := repository.NewEncounterRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// Initialize services
	personService := service.NewPersonService(personRepo, faceRepo)
	faceService := service.NewFaceService(faceRepo, personRepo)
	recognitionService := service.NewRecognitionService(faceRepo, personRepo, encounterRepo)
	encounterService := service.NewEncounterService(encounterRepo, personRepo)
	jobService := service.NewJobService(jobRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	personHandler := handler.NewPersonHandler(personService)
	faceHandler := handler.NewFaceHandler(faceService)
	recognitionHandler := handler.NewRecognitionHandler(recognitionService)
	encounterHandler := handler.NewEncounterHandler(encounterService)
	transcribeHandler := handler.NewTranscribeHandler(jobService)

	// Setup Gin router
	r := gin.Default()
	r.Use(middleware.RequestID())

	// API v1 routes
	v1 := r.Group("/v1")

	// Health check endpoint (no auth required)
	v1.GET("/healthz", healthHandler.Healthz)

	// Protected routes (require API key)
	protected := v1.Group("")
	protected.Use(middleware.APIKeyAuth())

	// Recognition endpoints
	protected.POST("/recognize", recognitionHandler.PostRecognize)

	// Person endpoints
	protected.GET("/persons", personHandler.ListPersons)
	protected.POST("/persons", personHandler.CreatePerson)
	protected.GET("/persons/:person_id", personHandler.GetPerson)
	protected.PATCH("/persons/:person_id", personHandler.UpdatePerson)
	protected.DELETE("/persons/:person_id", personHandler.DeletePerson)

	// Face endpoints
	protected.POST("/persons/:person_id/faces", faceHandler.AddFace)
	protected.GET("/persons/:person_id/faces", faceHandler.ListFaces)
	protected.DELETE("/persons/:person_id/faces/:face_id", faceHandler.DeleteFace)

	// Encounter endpoints
	protected.GET("/persons/:person_id/encounters", encounterHandler.ListEncounters)

	// Transcription endpoints
	protected.POST("/transcribe", transcribeHandler.PostTranscribe)
	protected.GET("/jobs/:job_id", transcribeHandler.GetJob)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
