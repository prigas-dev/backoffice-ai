package feature_generator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
	"github.com/prigas-dev/backoffice-ai/operations"

	_ "embed"
)

type IFeatureGenerator interface {
	GenerateFeature(ctx context.Context, prompt string) error
}

func NewReactFeatureGenerator(db *sql.DB, operationStore operations.IOperationStore, aiGenerator IAIGenerator) IFeatureGenerator {
	return &ReactFeatureGenerator{
		db:             db,
		operationStore: operationStore,
		aiGenerator:    aiGenerator,
	}
}

type ReactFeatureGenerator struct {
	db             *sql.DB
	operationStore operations.IOperationStore
	aiGenerator    IAIGenerator
}

func (g *ReactFeatureGenerator) GenerateFeature(ctx context.Context, prompt string) error {

	p, err := g.aiGenerator.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to create page component view: %w", err)
	}

	err = SaveFeatureToJsonFile(p)
	if err != nil {
		return fmt.Errorf("failed to save view json file: %w", err)
	}

	err = ComponentGenerator.GenerateComponentTSX(p.ReactComponent.TsxCode, "./http_server/public")
	if err != nil {
		return fmt.Errorf("failed to generate tsx component: %w", err)
	}

	for _, operation := range p.ServerOperations {
		err := g.operationStore.AddOperation(&operation)
		if err != nil {
			return fmt.Errorf("failed to store operation %s: %w", operation.Name, err)
		}
	}

	return nil

}

func SaveFeatureToJsonFile(p *Feature) error {
	outFile, err := os.Create(fmt.Sprintf("./AiGeneratedViews/%s.json", uuid.NewString()))
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
