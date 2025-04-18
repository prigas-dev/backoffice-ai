package AiAssistant

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
)

//go:embed system-instructions.txt
var systemInstructionsTmpl string

var ErrNoValidAnthropicResponse = errors.New("anthropic did not return any valid response")

func Assist(ctx context.Context, db *sql.DB, prompt string, databaseHints string) (*PageComponentView, error) {

	client := anthropic.NewClient()

	schema, err := PrintSchemaAsSQL(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get db schema: %w", err)
	}

	log.Println("Got schema from SQLite3 database")

	tmpl, err := template.New("system-instructions").Parse(systemInstructionsTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instructions template: %w", err)
	}

	log.Println("Parsed system-instructions template")

	var instructionsBuffer bytes.Buffer
	type TemplateData struct {
		DatabaseSchema string
		DatabaseHints  string
	}
	data := TemplateData{
		DatabaseSchema: schema,
		DatabaseHints:  databaseHints,
	}
	err = tmpl.Execute(&instructionsBuffer, data)
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

			errorStructure := &ErrorStructure{}
			err := json.Unmarshal([]byte(block.Text), errorStructure)
			if err == nil {
				if len(errorStructure.Error) > 0 {
					log.Println("Anthropic could not generate the view")
					return nil, fmt.Errorf("anthropic error: %s", errorStructure.Error)
				}
			}

			pageComponentView := &PageComponentView{}
			err = json.Unmarshal([]byte(block.Text), pageComponentView)
			if err == nil {
				log.Println("Successfully parsed anthropic response")
				return pageComponentView, nil
			}

			lastErr = fmt.Errorf("failed to parse anthropic response, %s: %w", block.Text, err)
		}
	}

	if lastErr == nil {
		return nil, ErrNoValidAnthropicResponse
	}

	return nil, lastErr
}

type QueryMode string

const (
	SingleRow    QueryMode = "single-row"
	MultipleRows QueryMode = "multiple-rows"
)

type Query struct {
	SQL       string    `json:"sql"`
	Mode      QueryMode `json:"mode"`      // "single-row" or "multiple-rows"
	MapToProp string    `json:"mapToProp"` // Name of the property to map the result to
}

type Component struct {
	ID   string `json:"id"`   // snake_case string identifier
	Code string `json:"code"` // TSX code
}

type PageComponentView struct {
	Queries   []Query   `json:"queries"`
	Component Component `json:"component"`
}

type ErrorStructure struct {
	Error string `json:"error"`
}
