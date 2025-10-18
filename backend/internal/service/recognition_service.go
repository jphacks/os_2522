package service

import (
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/jphacks/os_2522/backend/internal/repository"
	"github.com/jphacks/os_2522/backend/internal/utils"
)

// RecognitionService handles face recognition business logic
type RecognitionService struct {
	faceRepo      *repository.FaceRepository
	personRepo    *repository.PersonRepository
	encounterRepo *repository.EncounterRepository
}

// NewRecognitionService creates a new RecognitionService
func NewRecognitionService(
	faceRepo *repository.FaceRepository,
	personRepo *repository.PersonRepository,
	encounterRepo *repository.EncounterRepository,
) *RecognitionService {
	return &RecognitionService{
		faceRepo:      faceRepo,
		personRepo:    personRepo,
		encounterRepo: encounterRepo,
	}
}

// Recognize performs face recognition using client-provided embedding
func (s *RecognitionService) Recognize(req *models.RecognitionRequest) (*models.RecognitionResponse, error) {
	topK := req.TopK
	if topK == 0 {
		topK = 3 // Default
	}
	minScore := req.MinScore
	if minScore == 0 {
		minScore = 0.6 // Default
	}

	// Retrieve all face embeddings from database
	faceEntities, err := s.faceRepo.FindAllEmbeddings()
	if err != nil {
		return nil, err
	}

	// Calculate similarity scores
	type candidateScore struct {
		personID string
		faceID   string
		score    float64
	}

	var scores []candidateScore
	for _, faceEntity := range faceEntities {
		// Convert stored bytes back to float32
		storedEmbedding := utils.BytesToFloat32Slice(faceEntity.Embedding)
		if storedEmbedding == nil {
			continue // Skip invalid embeddings
		}

		// Calculate cosine similarity
		score := utils.CosineSimilarity(req.Embedding, storedEmbedding)
		if score >= minScore {
			scores = append(scores, candidateScore{
				personID: faceEntity.PersonID,
				faceID:   faceEntity.FaceID,
				score:    score,
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Limit to topK
	if len(scores) > topK {
		scores = scores[:topK]
	}

	// If no matches found
	if len(scores) == 0 {
		return &models.RecognitionResponse{
			Status:     models.RecognitionStatusUnknown,
			Candidates: []models.RecognitionCandidate{},
		}, nil
	}

	// Build candidates list
	candidates := make([]models.RecognitionCandidate, len(scores))
	for i, sc := range scores {
		person, err := s.personRepo.FindByID(sc.personID)
		if err != nil {
			continue
		}

		candidates[i] = models.RecognitionCandidate{
			PersonID:    person.PersonID,
			Name:        person.Name,
			Score:       sc.score,
			LastSummary: person.LastSummary,
		}
	}

	bestMatch := &candidates[0]

	return &models.RecognitionResponse{
		Status:     models.RecognitionStatusKnown,
		BestMatch:  bestMatch,
		Candidates: candidates,
	}, nil
}
