package service

import (
	"encoding/binary"
	"math"
	"mime/multipart"

	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
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

// Recognize performs face recognition
func (s *RecognitionService) Recognize(file *multipart.FileHeader, topK int, minScore float64) (*models.RecognitionResponse, error) {
	// TODO: Generate actual face embedding from uploaded image
	// For now, use a mock embedding
	queryEmbedding := generateMockEmbedding(512)

	// Retrieve all face embeddings from database
	faces, err := s.faceRepo.FindAllEmbeddings()
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
	for _, face := range faces {
		score := cosineSimilarity(queryEmbedding, face.Embedding)
		if score >= minScore {
			scores = append(scores, candidateScore{
				personID: face.PersonID,
				faceID:   face.FaceID,
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

// cosineSimilarity calculates cosine similarity between two embeddings
func cosineSimilarity(a, b []byte) float64 {
	if len(a) != len(b) {
		return 0
	}

	dim := len(a) / 4 // Assuming float32 (4 bytes each)
	var dotProduct, normA, normB float64

	for i := 0; i < dim; i++ {
		valA := math.Float32frombits(binary.LittleEndian.Uint32(a[i*4 : (i+1)*4]))
		valB := math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))

		dotProduct += float64(valA * valB)
		normA += float64(valA * valA)
		normB += float64(valB * valB)
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
