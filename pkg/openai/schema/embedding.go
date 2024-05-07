package schema

import "math"

///////////////////////////////////////////////////////////////////////////////
// TYPES

// An embedding object
type Embedding struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// An set of created embeddings
type Embeddings struct {
	Id    string      `json:"id"`
	Data  []Embedding `json:"data"`
	Model string      `json:"model"`
	Usage struct {
		PromptTokerns int `json:"prompt_tokens"`
		TotalTokens   int `json:"total_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e Embedding) CosineDistance(other Embedding) float64 {
	count := 0
	length_a := len(e.Embedding)
	length_b := len(other.Embedding)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(other.Embedding[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(e.Embedding[k], 2)
			continue
		}
		sumA += e.Embedding[k] * other.Embedding[k]
		s1 += math.Pow(e.Embedding[k], 2)
		s2 += math.Pow(other.Embedding[k], 2)
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2))
}
