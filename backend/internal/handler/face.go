package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/service"
)

// FaceHandler handles face management requests
type FaceHandler struct {
	faceService *service.FaceService
}

// NewFaceHandler creates a new FaceHandler
func NewFaceHandler(faceService *service.FaceService) *FaceHandler {
	return &FaceHandler{faceService: faceService}
}

// AddFace handles POST /persons/{person_id}/faces
func (h *FaceHandler) AddFace(c *gin.Context) {
	personID := c.Param("person_id")

	file, err := c.FormFile("image")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Image file is required"))
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		errors.RespondWithError(c, errors.UnsupportedMediaType("Only JPEG and PNG images are supported"))
		return
	}

	note := c.PostForm("note")
	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	face, err := h.faceService.AddFace(personID, file, notePtr)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, face)
}

// ListFaces handles GET /persons/{person_id}/faces
func (h *FaceHandler) ListFaces(c *gin.Context) {
	personID := c.Param("person_id")

	faceList, err := h.faceService.ListFaces(personID)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, faceList)
}

// DeleteFace handles DELETE /persons/{person_id}/faces/{face_id}
func (h *FaceHandler) DeleteFace(c *gin.Context) {
	personID := c.Param("person_id")
	faceID := c.Param("face_id")

	err := h.faceService.DeleteFace(personID, faceID)
	if err != nil {
		if err.Error() == "face not found" || err.Error() == "face does not belong to this person" {
			errors.RespondWithError(c, errors.NotFound("Face not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
