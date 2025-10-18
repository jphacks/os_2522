package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/database"
	"github.com/jphacks/os_2522/backend/internal/handler"
	"github.com/jphacks/os_2522/backend/internal/middleware"
	"github.com/jphacks/os_2522/backend/internal/repository"
	"github.com/jphacks/os_2522/backend/internal/service"
	"github.com/jphacks/os_2522/backend/internal/utils"
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

	// Initialize Gemini client
	ctx := context.Background()
	geminiClient, err := utils.NewGeminiClient(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize Gemini client: %v", err)
		log.Println("Summarization endpoint will not be available")
	}
	defer func() {
		if geminiClient != nil {
			geminiClient.Close()
		}
	}()

	// Initialize repositories
	personRepo := repository.NewPersonRepository(db)
	faceRepo := repository.NewFaceRepository(db)
	encounterRepo := repository.NewEncounterRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// Initialize services
	personService := service.NewPersonService(personRepo, faceRepo)
	faceService := service.NewFaceService(faceRepo, personRepo)
	faceExtractionService := service.NewFaceExtractionService()
	recognitionService := service.NewRecognitionService(faceRepo, personRepo, encounterRepo)
	encounterService := service.NewEncounterService(encounterRepo, personRepo)
	jobService := service.NewJobService(jobRepo)
	var summarizeService *service.SummarizeService
	if geminiClient != nil {
		summarizeService = service.NewSummarizeService(geminiClient)
	}

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	personHandler := handler.NewPersonHandler(personService)
	faceHandler := handler.NewFaceHandler(faceService, faceExtractionService)
	recognitionHandler := handler.NewRecognitionHandler(recognitionService, faceExtractionService)
	encounterHandler := handler.NewEncounterHandler(encounterService)
	transcribeHandler := handler.NewTranscribeHandler(jobService)
	var summarizeHandler *handler.SummarizeHandler
	if summarizeService != nil {
		summarizeHandler = handler.NewSummarizeHandler(summarizeService)
	}

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
	protected.POST("/recognize-image", recognitionHandler.PostRecognizeImage)

	// Person endpoints
	protected.GET("/persons", personHandler.ListPersons)
	protected.POST("/persons", personHandler.CreatePerson)
	protected.GET("/persons/:person_id", personHandler.GetPerson)
	protected.PATCH("/persons/:person_id", personHandler.UpdatePerson)
	protected.DELETE("/persons/:person_id", personHandler.DeletePerson)

	// Face endpoints
	protected.POST("/persons/:person_id/faces", faceHandler.AddFace)
	protected.POST("/persons/:person_id/faces-image", faceHandler.AddFaceImage)
	protected.GET("/persons/:person_id/faces", faceHandler.ListFaces)
	protected.DELETE("/persons/:person_id/faces/:face_id", faceHandler.DeleteFace)

	// Encounter endpoints
	protected.GET("/persons/:person_id/encounters", encounterHandler.ListEncounters)

	// Transcription endpoints
	protected.POST("/transcribe", transcribeHandler.PostTranscribe)
	protected.GET("/jobs/:job_id", transcribeHandler.GetJob)

	// Summarization endpoint
	if summarizeHandler != nil {
		protected.POST("/summarize", summarizeHandler.PostSummarize)
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
