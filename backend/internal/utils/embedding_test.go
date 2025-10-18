package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat32SliceToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected int // Expected byte length
	}{
		{
			name:     "empty slice",
			input:    []float32{},
			expected: 0,
		},
		{
			name:     "single value",
			input:    []float32{1.5},
			expected: 4, // 1 float32 = 4 bytes
		},
		{
			name:     "multiple values",
			input:    []float32{1.0, 2.0, 3.0},
			expected: 12, // 3 float32s = 12 bytes
		},
		{
			name:     "512-dimensional embedding",
			input:    make([]float32, 512),
			expected: 2048, // 512 * 4 = 2048 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float32SliceToBytes(tt.input)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestBytesToFloat32Slice(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []float32
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: []float32{},
		},
		{
			name:     "invalid byte length (not multiple of 4)",
			input:    []byte{0, 0, 0}, // 3 bytes, not divisible by 4
			expected: nil,
		},
		{
			name:     "single float32",
			input:    Float32SliceToBytes([]float32{1.5}),
			expected: []float32{1.5},
		},
		{
			name:     "multiple float32s",
			input:    Float32SliceToBytes([]float32{1.0, 2.0, 3.0}),
			expected: []float32{1.0, 2.0, 3.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToFloat32Slice(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, len(tt.expected), len(result))
				for i := range tt.expected {
					assert.InDelta(t, tt.expected[i], result[i], 0.0001)
				}
			}
		})
	}
}

func TestFloat32SliceToBytes_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []float32
	}{
		{
			name:  "positive values",
			input: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		},
		{
			name:  "negative values",
			input: []float32{-0.1, -0.2, -0.3, -0.4, -0.5},
		},
		{
			name:  "mixed values",
			input: []float32{-1.5, 0.0, 1.5, 2.5, -2.5},
		},
		{
			name:  "special values",
			input: []float32{0.0, -0.0, 1.0, -1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to bytes
			bytes := Float32SliceToBytes(tt.input)

			// Convert back to float32
			result := BytesToFloat32Slice(bytes)

			// Verify round-trip conversion
			assert.Equal(t, len(tt.input), len(result))
			for i := range tt.input {
				assert.InDelta(t, tt.input[i], result[i], 0.0001)
			}
		})
	}
}

func TestCalculateEmbeddingChecksum(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected string // We'll check the format, not the exact hash
	}{
		{
			name:  "simple embedding",
			input: []float32{0.5, 0.5, 0.5},
		},
		{
			name:  "512-dimensional embedding",
			input: make([]float32, 512),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEmbeddingChecksum(tt.input)

			// Check format: should start with "sha256:"
			assert.Contains(t, result, "sha256:")

			// Check that the hash part is 64 characters (SHA256 hex digest)
			hashPart := result[7:] // Skip "sha256:"
			assert.Equal(t, 64, len(hashPart))

			// Same input should always produce same hash
			result2 := CalculateEmbeddingChecksum(tt.input)
			assert.Equal(t, result, result2)
		})
	}
}

func TestCalculateEmbeddingChecksum_Uniqueness(t *testing.T) {
	// Different embeddings should produce different checksums
	embedding1 := []float32{0.1, 0.2, 0.3}
	embedding2 := []float32{0.1, 0.2, 0.4} // Slightly different

	checksum1 := CalculateEmbeddingChecksum(embedding1)
	checksum2 := CalculateEmbeddingChecksum(embedding2)

	assert.NotEqual(t, checksum1, checksum2)
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.0001,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{-1.0, 0.0, 0.0},
			expected: -1.0,
			delta:    0.0001,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 0.0,
			delta:    0.0001,
		},
		{
			name:     "similar vectors (45 degrees)",
			a:        []float32{1.0, 1.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0 / math.Sqrt(2), // cos(45°) ≈ 0.707
			delta:    0.01,
		},
		{
			name:     "normalized embeddings",
			a:        []float32{0.6, 0.8},
			b:        []float32{0.6, 0.8},
			expected: 1.0,
			delta:    0.0001,
		},
		{
			name:     "different length vectors",
			a:        []float32{1.0, 2.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 0.0, // Should return 0 for mismatched lengths
			delta:    0.0001,
		},
		{
			name:     "empty vectors",
			a:        []float32{},
			b:        []float32{},
			expected: 0.0,
			delta:    0.0001,
		},
		{
			name:     "zero vector",
			a:        []float32{0.0, 0.0, 0.0},
			b:        []float32{1.0, 1.0, 1.0},
			expected: 0.0, // Zero norm
			delta:    0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func TestCosineSimilarity_Symmetry(t *testing.T) {
	// Cosine similarity should be symmetric: sim(a, b) == sim(b, a)
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{4.0, 5.0, 6.0}

	sim1 := CosineSimilarity(a, b)
	sim2 := CosineSimilarity(b, a)

	assert.InDelta(t, sim1, sim2, 0.0001)
}

func TestCosineSimilarity_Range(t *testing.T) {
	// Cosine similarity should always be in range [-1, 1]
	tests := []struct {
		name string
		a    []float32
		b    []float32
	}{
		{
			name: "random positive values",
			a:    []float32{0.5, 0.3, 0.8, 0.1},
			b:    []float32{0.2, 0.9, 0.4, 0.7},
		},
		{
			name: "random mixed values",
			a:    []float32{-0.5, 0.3, -0.8, 0.1},
			b:    []float32{0.2, -0.9, 0.4, -0.7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			assert.GreaterOrEqual(t, result, -1.0)
			assert.LessOrEqual(t, result, 1.0)
		})
	}
}

func BenchmarkFloat32SliceToBytes(b *testing.B) {
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Float32SliceToBytes(embedding)
	}
}

func BenchmarkBytesToFloat32Slice(b *testing.B) {
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = 0.5
	}
	bytes := Float32SliceToBytes(embedding)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BytesToFloat32Slice(bytes)
	}
}

func BenchmarkCosineSimilarity(b *testing.B) {
	a := make([]float32, 512)
	bb := make([]float32, 512)
	for i := range a {
		a[i] = 0.5
		bb[i] = 0.6
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(a, bb)
	}
}

func BenchmarkCalculateEmbeddingChecksum(b *testing.B) {
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = 0.5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateEmbeddingChecksum(embedding)
	}
}
