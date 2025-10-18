package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/handler"
	"github.com/jphacks/os_2522/backend/internal/middleware"
)

func main() {
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.RequestID())

	// API v1 routes
	v1 := r.Group("/v1")

	// Health check endpoint (no auth required)
	healthHandler := handler.NewHealthHandler()
	v1.GET("/healthz", healthHandler.Healthz)

	// Protected routes (require API key)
	protected := v1.Group("")
	protected.Use(middleware.APIKeyAuth())

	// Recognition endpoints
	recognitionHandler := handler.NewRecognitionHandler()
	protected.POST("/recognize", recognitionHandler.PostRecognize)

	// Person endpoints
	personHandler := handler.NewPersonHandler()
	protected.GET("/persons", personHandler.ListPersons)
	protected.POST("/persons", personHandler.CreatePerson)
	protected.GET("/persons/:person_id", personHandler.GetPerson)
	protected.PATCH("/persons/:person_id", personHandler.UpdatePerson)
	protected.DELETE("/persons/:person_id", personHandler.DeletePerson)

	// Face endpoints
	faceHandler := handler.NewFaceHandler()
	protected.POST("/persons/:person_id/faces", faceHandler.AddFace)
	protected.GET("/persons/:person_id/faces", faceHandler.ListFaces)
	protected.DELETE("/persons/:person_id/faces/:face_id", faceHandler.DeleteFace)

	// Encounter endpoints
	encounterHandler := handler.NewEncounterHandler()
	protected.GET("/persons/:person_id/encounters", encounterHandler.ListEncounters)

	// Transcription endpoints
	transcribeHandler := handler.NewTranscribeHandler()
	protected.POST("/transcribe", transcribeHandler.PostTranscribe)
	protected.GET("/jobs/:job_id", transcribeHandler.GetJob)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
