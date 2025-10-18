package service

import (
	"encoding/binary"
	"fmt"
	"math"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
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

// AddFace adds a new face to a person
func (s *FaceService) AddFace(personID string, file *multipart.FileHeader, note *string) (*models.Face, error) {
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

	// TODO: Generate actual face embedding from image
	// For now, generate a mock embedding (512-dimensional)
	embedding := generateMockEmbedding(512)

	entity := &repository.FaceEntity{
		FaceID:       faceID,
		PersonID:     personID,
		Embedding:    embedding,
		EmbeddingDim: 512,
		Note:         note,
		CreatedAt:    time.Now(),
	}

	if err := s.faceRepo.Create(entity); err != nil {
		return nil, err
	}

	return &models.Face{
		FaceID:       entity.FaceID,
		PersonID:     entity.PersonID,
		ImageURL:     entity.ImagePath,
		EmbeddingDim: entity.EmbeddingDim,
		Note:         entity.Note,
		CreatedAt:    entity.CreatedAt,
	}, nil
}

// ListFaces retrieves all faces for a person
func (s *FaceService) ListFaces(personID string) (*models.FaceList, error) {
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
		faces[i] = models.Face{
			FaceID:       entity.FaceID,
			PersonID:     entity.PersonID,
			ImageURL:     entity.ImagePath,
			EmbeddingDim: entity.EmbeddingDim,
			Note:         entity.Note,
			CreatedAt:    entity.CreatedAt,
		}
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

// generateMockEmbedding generates a mock embedding vector
func generateMockEmbedding(dim int) []byte {
	// Create a normalized random vector
	buf := make([]byte, dim*4) // 4 bytes per float32
	for i := 0; i < dim; i++ {
		val := float32(0.5) // Mock value
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(val))
	}
	return buf
}
