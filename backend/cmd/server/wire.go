package main

import (
	"log"

	"github.com/jphacks/os_2522/backend/internal/database"
	"github.com/jphacks/os_2522/backend/internal/handler"
	"github.com/jphacks/os_2522/backend/internal/repository"
	"github.com/jphacks/os_2522/backend/internal/service"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	Health      *handler.HealthHandler
	Person      *handler.PersonHandler
	Face        *handler.FaceHandler
	Recognition *handler.RecognitionHandler
	Encounter   *handler.EncounterHandler
	Transcribe  *handler.TranscribeHandler
}

func main() {
	if _, err := InitializeApp(); err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}

// InitializeApp initializes the application dependencies
func InitializeApp() (*Handlers, error) {
	// Initialize database
	dbConfig := database.GetConfigFromEnv()
	db, err := database.NewDB(dbConfig)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		return nil, err
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
	handlers := &Handlers{
		Health:      handler.NewHealthHandler(),
		Person:      handler.NewPersonHandler(personService),
		Face:        handler.NewFaceHandler(faceService),
		Recognition: handler.NewRecognitionHandler(recognitionService),
		Encounter:   handler.NewEncounterHandler(encounterService),
		Transcribe:  handler.NewTranscribeHandler(jobService),
	}

	return handlers, nil
}
