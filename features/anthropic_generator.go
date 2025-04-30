package features

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	_ "embed"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/prigas-dev/backoffice-ai/features/instruction_files"
	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/prigas-dev/backoffice-ai/utils"
)

type IAIGenerator interface {
	Generate(ctx context.Context, prompt string) (*Feature, error)
}

type Feature struct {
	Name             string                  `json:"name"`
	Label            string                  `json:"label"`
	Description      string                  `json:"description"`
	ReactComponent   *ReactComponent         `json:"reactComponent"`
	ServerOperations []*operations.Operation `json:"serverOperations"`
}

type ReactComponent struct {
	TsxCode string `json:"tsxCode"`
}

func NewAIGenerator(db *sql.DB, databaseSchemaGenerator IDatabaseSchemaGenerator, instructionsTemplateData *InstructionsTemplateData) IAIGenerator {
	return &AnthropicGenerator{
		db:                      db,
		databaseSchemaGenerator: databaseSchemaGenerator,
		instructionsTemplateData: &AnthropicInstructionsTemplateData{
			SystemName:        instructionsTemplateData.SystemName,
			SystemDescription: instructionsTemplateData.SystemDescription,
			DatabaseEngine:    instructionsTemplateData.DatabaseEngine,
			DatabaseHints:     instructionsTemplateData.DatabaseHints,
			FeatureJSONSchema: instruction_files.FeatureJSONSchema.Content,
			ErrorJSONSchema:   instruction_files.ErrorJSONSchema.Content,
			ValidFeatureJSON:  instruction_files.ExampleFeatureJSON.Content,
			ValidFeatureFiles: instruction_files.ExampleFeatureFiles,
		},
	}
}

type InstructionsTemplateData struct {
	SystemName        string
	SystemDescription string
	DatabaseEngine    string
	DatabaseHints     string
}

type AnthropicGenerator struct {
	db                       *sql.DB
	databaseSchemaGenerator  IDatabaseSchemaGenerator
	instructionsTemplateData *AnthropicInstructionsTemplateData
}

type AnthropicInstructionsTemplateData struct {
	SystemName        string
	SystemDescription string
	DatabaseEngine    string
	DatabaseHints     string

	DatabaseSchema    string
	ErrorJSONSchema   string
	FeatureJSONSchema string
	ValidFeatureJSON  string
	ValidFeatureFiles []instruction_files.File
}

func (g *AnthropicGenerator) Generate(ctx context.Context, prompt string) (*Feature, error) {

	client := anthropic.NewClient()

	schema, err := g.databaseSchemaGenerator.GenerateSchemaSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to get db schema: %w", err)
	}

	log.Println("Got schema from SQLite3 database")

	g.instructionsTemplateData.DatabaseSchema = schema
	instructions, err := utils.DoTemplate("system-instructions", instruction_files.SystemInstructionsTemplate.Content, g.instructionsTemplateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate instructions: %w", err)
	}

	log.Println("Executed template successfuly")

	messages := []anthropic.MessageParam{
		{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				anthropic.ContentBlockParamOfRequestTextBlock(prompt),
			},
		},
	}

	anthropicResponse, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:       anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens:   10_000,
		Temperature: anthropic.Float(0.5),
		System: []anthropic.TextBlockParam{
			{
				Text: instructions,
			},
		},
		Messages: messages,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to anthropic: %w", err)
	}

	log.Println("Got response from anthropic")

	var lastErr error = nil

	for _, block := range anthropicResponse.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:

			log.Println(block.Text)

			errorStructure := &AIGenerationError{}
			err := json.Unmarshal([]byte(block.Text), errorStructure)
			if err == nil {
				if len(errorStructure.Error) > 0 {
					log.Println("Anthropic could not generate the view")
					return nil, fmt.Errorf("anthropic error: %s", errorStructure.Error)
				}
			}

			feature := &Feature{}
			err = json.Unmarshal([]byte(block.Text), feature)
			if err == nil {
				log.Println("Successfully parsed anthropic response")
				return feature, nil
			}

			lastErr = fmt.Errorf("failed to parse anthropic response, %s: %w", block.Text, err)
		}
	}

	if lastErr == nil {
		return nil, ErrNoValidAnthropicResponse
	}

	return nil, lastErr
}

var ErrNoValidAnthropicResponse = errors.New("anthropic did not return any valid response")

type AIGenerationError struct {
	Error string `json:"error"`
}
