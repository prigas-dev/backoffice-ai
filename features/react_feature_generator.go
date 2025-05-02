package features

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/prigas-dev/backoffice-ai/frontend"

	_ "embed"
)

type IFeatureGenerator interface {
	GenerateFeature(ctx context.Context, prompt string, featureContext *Feature) (*Feature, error)
}

func NewReactFeatureGenerator(db *sql.DB, featureStore IFeatureStore, aiGenerator IAIGenerator, frontendBuilder frontend.IBuilder) IFeatureGenerator {
	return &ReactFeatureGenerator{
		db:              db,
		featureStore:    featureStore,
		aiGenerator:     aiGenerator,
		frontendBuilder: frontendBuilder,
	}
}

type ReactFeatureGenerator struct {
	db              *sql.DB
	featureStore    IFeatureStore
	aiGenerator     IAIGenerator
	frontendBuilder frontend.IBuilder
}

func (g *ReactFeatureGenerator) GenerateFeature(ctx context.Context, prompt string, featureContext *Feature) (*Feature, error) {

	feature, err := g.aiGenerator.Generate(ctx, prompt, featureContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create page component view: %w", err)
	}

	err = SaveFeatureToJsonFile(feature)
	if err != nil {
		return nil, fmt.Errorf("failed to save view json file: %w", err)
	}

	err = g.featureStore.AddFeature(feature)
	if err != nil {
		return nil, fmt.Errorf("failed to store feature %s: %w", feature.Name, err)
	}

	err = g.frontendBuilder.BuildFrontend()
	if err != nil {
		return nil, fmt.Errorf("failed to build frontend: %w", err)
	}

	return feature, nil
}

func SaveFeatureToJsonFile(p *Feature) error {
	outFile, err := os.Create(fmt.Sprintf("./AiGeneratedViews/%s.json", p.Name))
	if err != nil {
		return fmt.Errorf("failed to create view json file: %w", err)
	}
	defer outFile.Close()

	formattedJson, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format json: %w", err)
	}

	_, err = outFile.Write(formattedJson)
	if err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}
