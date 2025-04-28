package feature_generator

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"text/template"

	_ "embed"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/prigas-dev/backoffice-ai/feature_generator/instruction_files"
	"github.com/prigas-dev/backoffice-ai/operations"
)

type IAIGenerator interface {
	Generate(ctx context.Context, prompt string) (*Feature, error)
}

type Feature struct {
	ReactComponent   ReactComponent         `json:"reactComponent"`
	ServerOperations []operations.Operation `json:"serverOperations"`
}

type ReactComponent struct {
	TsxCode string `json:"tsxCode"`
}

func NewAIGenerator(db *sql.DB, databaseSchemaGenerator IDatabaseSchemaGenerator, instructionsTemplateData *InstructionsTemplateData) IAIGenerator {
	return &AnthropicGenerator{
		db:                       db,
		databaseSchemaGenerator:  databaseSchemaGenerator,
		instructionsTemplateData: instructionsTemplateData,
	}
}

type AnthropicGenerator struct {
	db                       *sql.DB
	databaseSchemaGenerator  IDatabaseSchemaGenerator
	instructionsTemplateData *InstructionsTemplateData
}

type InstructionsTemplateData struct {
	SystemName        string
	SystemDescription string
	DatabaseEngine    string
	DatabaseSchema    string
	DatabaseHints     string
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

	tmpl, err := template.New("system-instructions").Parse(instruction_files.SystemInstructionsTemplate.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instructions template: %w", err)
	}

	log.Println("Parsed system-instructions template")

	var instructionsBuffer bytes.Buffer

	g.instructionsTemplateData.DatabaseSchema = schema

	err = tmpl.Execute(&instructionsBuffer, g.instructionsTemplateData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute instructions template: %w", err)
	}

	instructions := instructionsBuffer.String()

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
