package repository

import (
	"gorm.io/gorm"
)

// FaceRepository handles face data access
type FaceRepository struct {
	db *gorm.DB
}

// NewFaceRepository creates a new FaceRepository
func NewFaceRepository(db *gorm.DB) *FaceRepository {
	return &FaceRepository{db: db}
}

// Create creates a new face
func (r *FaceRepository) Create(face *FaceEntity) error {
	return r.db.Create(face).Error
}

// FindByPersonID retrieves all faces for a person
func (r *FaceRepository) FindByPersonID(personID string) ([]FaceEntity, error) {
	var faces []FaceEntity
	err := r.db.Where("person_id = ?", personID).Order("created_at DESC").Find(&faces).Error
	return faces, err
}

// FindByID retrieves a face by ID
func (r *FaceRepository) FindByID(faceID string) (*FaceEntity, error) {
	var face FaceEntity
	if err := r.db.First(&face, "face_id = ?", faceID).Error; err != nil {
		return nil, err
	}
	return &face, nil
}

// Delete soft deletes a face
func (r *FaceRepository) Delete(faceID string) error {
	return r.db.Delete(&FaceEntity{}, "face_id = ?", faceID).Error
}

// FindAllEmbeddings retrieves all face embeddings for similarity search
func (r *FaceRepository) FindAllEmbeddings() ([]FaceEntity, error) {
	var faces []FaceEntity
	err := r.db.Select("face_id, person_id, embedding, embedding_dim").Find(&faces).Error
	return faces, err
}
