package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// Float32SliceToBytes converts a []float32 to a []byte using little-endian encoding
func Float32SliceToBytes(floats []float32) []byte {
	buf := make([]byte, len(floats)*4) // 4 bytes per float32
	for i, val := range floats {
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(val))
	}
	return buf
}

// BytesToFloat32Slice converts a []byte to a []float32 using little-endian encoding
func BytesToFloat32Slice(buf []byte) []float32 {
	if len(buf)%4 != 0 {
		return nil // Invalid buffer length
	}
	dim := len(buf) / 4
	floats := make([]float32, dim)
	for i := 0; i < dim; i++ {
		bits := binary.LittleEndian.Uint32(buf[i*4 : (i+1)*4])
		floats[i] = math.Float32frombits(bits)
	}
	return floats
}

// CalculateEmbeddingChecksum calculates SHA256 checksum of an embedding vector
func CalculateEmbeddingChecksum(embedding []float32) string {
	bytes := Float32SliceToBytes(embedding)
	hash := sha256.Sum256(bytes)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// CosineSimilarity calculates cosine similarity between two embedding vectors
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
