package AiAssistant

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	_ "embed"

	"github.com/anthropics/anthropic-sdk-go"
)

//go:embed system-instructions.txt
var systemInstructionsTmpl string

var ErrNoValidAnthropicResponse = errors.New("anthropic did not return any valid response")

func Assist(ctx context.Context, db *sql.DB, prompt string) (*PageComponentView, error) {

	client := anthropic.NewClient()

	schema, err := PrintSchemaAsSQL(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get db schema: %w", err)
	}

	tmpl, err := template.New("system-instructions").Parse(systemInstructionsTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instructions template: %w", err)
	}

	var instructionsBuffer bytes.Buffer
	type TemplateData struct {
		DatabaseSchema string
	}
	data := TemplateData{
		DatabaseSchema: schema,
	}
	err = tmpl.Execute(&instructionsBuffer, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute instructions template: %w", err)
	}

	instructions := instructionsBuffer.String()

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

	var lastErr error = nil

	for _, block := range anthropicResponse.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:
			pageComponentView := &PageComponentView{}
			err := json.Unmarshal([]byte(block.Text), pageComponentView)
			if err == nil {
				// TODO handle error response case scenario
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
	ID   string `json:"id"`   // Snake-case string identifier
	Code string `json:"code"` // TSX code
}

type PageComponentView struct {
	Queries   []Query   `json:"queries"`
	Component Component `json:"component"`
}
