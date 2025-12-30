package config

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type RecommenderModel struct {
	Features      []string           `json:"features"`
	Coefficients  [][]float64        `json:"coefficients"`
	Intercept     []float64          `json:"intercept"`
	Encoders      map[string]map[string]int `json:"encoders"`
}

func LoadRecommender(workDir string) (*RecommenderModel, error) {
	modelPath := filepath.Join(workDir, "ml", "recommender", "model.json")
	
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("recommender model not found: %w (run 'gptcode ml train recommender')", err)
	}
	
	var model RecommenderModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse recommender model: %w", err)
	}
	
	return &model, nil
}

func (r *RecommenderModel) extractModelFeatures(modelID string) (hasCoder, hasInstant, modelSize int) {
	modelLower := strings.ToLower(modelID)
	
	if strings.Contains(modelLower, "coder") || strings.Contains(modelLower, "code") {
		hasCoder = 1
	}
	
	if strings.Contains(modelLower, "instant") || strings.Contains(modelLower, "flash") {
		hasInstant = 1
	}
	
	sizes := []string{"405b", "120b", "70b", "32b", "33b", "22b", "9b", "8b", "3b"}
	for _, size := range sizes {
		if strings.Contains(modelLower, size) {
			fmt.Sscanf(size, "%db", &modelSize)
			break
		}
	}
	
	return
}

func (r *RecommenderModel) PredictSuccess(
	modelID string,
	action ActionType,
	language string,
	complexity string,
	contextWindow int,
	costPer1M float64,
) float64 {
	actionEncoded := 0
	if val, ok := r.Encoders["action"][string(action)]; ok {
		actionEncoded = val
	}
	
	languageEncoded := 0
	if val, ok := r.Encoders["language"][strings.ToLower(language)]; ok {
		languageEncoded = val
	}
	
	complexityEncoded := 0
	if val, ok := r.Encoders["complexity"][complexity]; ok {
		complexityEncoded = val
	}
	
	hasCoder, hasInstant, modelSize := r.extractModelFeatures(modelID)
	
	logCost := math.Log1p(costPer1M)
	logContext := math.Log1p(float64(contextWindow))
	
	features := []float64{
		float64(actionEncoded),
		float64(languageEncoded),
		float64(complexityEncoded),
		logCost,
		logContext,
		float64(hasCoder),
		float64(hasInstant),
		float64(modelSize),
	}
	
	if len(r.Coefficients) == 0 || len(r.Coefficients[0]) != len(features) {
		return 0.5
	}
	
	logit := r.Intercept[0]
	for i, feat := range features {
		logit += r.Coefficients[0][i] * feat
	}
	
	prob := 1.0 / (1.0 + math.Exp(-logit))
	return prob
}
