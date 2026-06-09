package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// FakeClient produces deterministic 1024-dim unit vectors from the SHA-256 of
// each input text. Same text → same vector across runs, different texts →
// different vectors. Cosine similarity between two texts is then a stable
// function of the bytes alone — useful for wiring up the search path locally
// without paying for real embeddings.
//
// NOT suitable for production: there is no semantic signal in the vectors.
// Use Voyage or OpenAI in real environments.
type FakeClient struct{}

func NewFakeClient() *FakeClient { return &FakeClient{} }

func (FakeClient) Name() string { return "fake:sha256-unit-vector" }

func (FakeClient) Embed(_ context.Context, texts []string, _ string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = vectorFromHash(t)
	}
	return out, nil
}

// vectorFromHash expands a SHA-256 digest into a Dim-length unit vector. The
// digest is interpreted as a stream of int32 little-endian values, normalized
// to [-1, 1]; we recycle the hash by re-hashing on overflow so the produced
// vector is fully deterministic but covers the full embedding width.
func vectorFromHash(s string) []float32 {
	vec := make([]float32, Dim)
	seed := sha256.Sum256([]byte(s))
	buf := seed[:]
	cursor := 0
	for i := 0; i < Dim; i++ {
		if cursor+4 > len(buf) {
			next := sha256.Sum256(buf)
			buf = append(buf, next[:]...)
		}
		n := int32(binary.LittleEndian.Uint32(buf[cursor : cursor+4]))
		cursor += 4
		vec[i] = float32(n) / float32(math.MaxInt32)
	}
	return normalize(vec)
}

func normalize(v []float32) []float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return v
	}
	inv := float32(1.0 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
	return v
}
