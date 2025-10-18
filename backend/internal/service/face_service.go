package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/jphacks/os_2522/backend/internal/repository"
	"github.com/jphacks/os_2522/backend/internal/utils"
	"gorm.io/gorm"
)

// FaceService handles face business logic
type FaceService struct {
	faceRepo   *repository.FaceRepository
	personRepo *repository.PersonRepository
}

// NewFaceService creates a new FaceService
func NewFaceService(faceRepo *repository.FaceRepository, personRepo *repository.PersonRepository) *FaceService {
	return &FaceService{
		faceRepo:   faceRepo,
		personRepo: personRepo,
	}
}

// AddFace adds a new face to a person with client-provided embedding
func (s *FaceService) AddFace(personID string, req *models.FaceEmbeddingRequest) (*models.Face, error) {
	// Verify person exists
	_, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	// Generate face ID
	faceID := fmt.Sprintf("f-%s", uuid.New().String()[:8])

	// Convert float32 slice to bytes
	embeddingBytes := utils.Float32SliceToBytes(req.Embedding)

	// Calculate checksum
	checksum := utils.CalculateEmbeddingChecksum(req.Embedding)

	entity := &repository.FaceEntity{
		FaceID:            faceID,
		PersonID:          personID,
		Embedding:         embeddingBytes,
		EmbeddingDim:      req.EmbeddingDim,
		ModelVersion:      &req.ModelVersion,
		EmbeddingChecksum: &checksum,
		SourceImageHash:   req.SourceImageHash,
		Note:              req.Note,
		CreatedAt:         time.Now(),
	}

	if err := s.faceRepo.Create(entity); err != nil {
		return nil, err
	}

	return &models.Face{
		FaceID:            entity.FaceID,
		PersonID:          entity.PersonID,
		ImageURL:          entity.ImagePath,
		EmbeddingDim:      entity.EmbeddingDim,
		ModelVersion:      entity.ModelVersion,
		EmbeddingChecksum: entity.EmbeddingChecksum,
		Note:              entity.Note,
		CreatedAt:         entity.CreatedAt,
	}, nil
}

// ListFaces retrieves all faces for a person
func (s *FaceService) ListFaces(personID string, includeEmbedding bool) (*models.FaceList, error) {
	// Verify person exists
	_, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	entities, err := s.faceRepo.FindByPersonID(personID)
	if err != nil {
		return nil, err
	}

	faces := make([]models.Face, len(entities))
	for i, entity := range entities {
		face := models.Face{
			FaceID:            entity.FaceID,
			PersonID:          entity.PersonID,
			ImageURL:          entity.ImagePath,
			EmbeddingDim:      entity.EmbeddingDim,
			ModelVersion:      entity.ModelVersion,
			EmbeddingChecksum: entity.EmbeddingChecksum,
			Note:              entity.Note,
			CreatedAt:         entity.CreatedAt,
		}

		// Include embedding vector if requested
		if includeEmbedding && entity.Embedding != nil {
			face.Embedding = utils.BytesToFloat32Slice(entity.Embedding)
		}

		faces[i] = face
	}

	return &models.FaceList{
		Items: faces,
	}, nil
}

// DeleteFace deletes a face
func (s *FaceService) DeleteFace(personID, faceID string) error {
	// Verify face exists and belongs to person
	face, err := s.faceRepo.FindByID(faceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("face not found")
		}
		return err
	}

	if face.PersonID != personID {
		return fmt.Errorf("face does not belong to this person")
	}

	return s.faceRepo.Delete(faceID)
}
