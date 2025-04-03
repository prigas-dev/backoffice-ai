package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
	"github.com/joho/godotenv"
)

func Sqliter(db *sql.DB, prompt string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := anthropic.NewClient()

	schema, err := PrintSchemaAsSQL(db)
	if err != nil {
		log.Fatalf("Failed to get db schema: %v", err)
	}

	messages := []anthropic.MessageParam{
		{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				anthropic.ContentBlockParamOfRequestTextBlock(prompt),
			},
		},
	}

	tools := []anthropic.ToolUnionParam{
		anthropic.ToolUnionParamOfTool(QuerySelectInputSchema, "query_select"),
	}

	for {
		message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
			Model:       anthropic.ModelClaude3_7SonnetLatest,
			MaxTokens:   10_000,
			Temperature: anthropic.Float(0.5),
			System: []anthropic.TextBlockParam{
				{
					Text: "You are a business analyst. I'm going to ask you some questions, and you have to answer me based on queries on a SQLite database. Here is the schema you can base on:\n" + schema,
				},
			},
			Messages: messages,
			Tools:    tools,
		})

		if err != nil {
			panic(err)
		}

		print(color("[assistant]: "))
		for _, block := range message.Content {
			switch block := block.AsAny().(type) {
			case anthropic.TextBlock:
				println(block.Text)
				println()
			case anthropic.ToolUseBlock:
				inputJSON, _ := json.Marshal(block.Input)
				println(block.Name + ": " + string(inputJSON))
				println()
			}
		}

		messages = append(messages, message.ToParam())
		toolResults := []anthropic.ContentBlockParamUnion{}

		for _, block := range message.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.ToolUseBlock:
				print(color("[user (" + block.Name + ")]: "))

				var response interface{}
				switch block.Name {
				case "query_select":
					var input QuerySelectInput

					err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
					if err != nil {
						panic(err)
					}

					response, err = QuerySelect(db, input)
					if err != nil {
						response = QuerySelectError{
							SqlError: err.Error(),
						}
					}
				}

				b, err := json.Marshal(response)
				if err != nil {
					panic(err)
				}

				println(string(b))

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), false))
			}

		}
		if len(toolResults) == 0 {
			break
		}
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}

type QuerySelectInput struct {
	Query string `json:"query" jsonschema_description:"SQLite SELECT query to run. You can only query on tables and columns from the schema I just provided to you."`
}

var QuerySelectInputSchema = GenerateSchema[QuerySelectInput]()

type QuerySelectError struct {
	SqlError string `json:"sqlError"`
}

type QuerySelectOutput struct {
	Rows [][]any `json:"rows"`
}

func QuerySelect(db *sql.DB, input QuerySelectInput) (QuerySelectOutput, error) {
	rows, err := db.Query(input.Query)
	if err != nil {
		return QuerySelectOutput{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return QuerySelectOutput{}, err
	}

	scannedRows := [][]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			return QuerySelectOutput{}, err
		}

		for i, val := range values {
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}
		}

		scannedRows = append(scannedRows, values)
	}
	err = rows.Err()
	if err != nil {
		return QuerySelectOutput{}, err
	}

	output := QuerySelectOutput{
		Rows: scannedRows,
	}

	return output, nil
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

func color(s string) string {
	return fmt.Sprintf("\033[1;%sm%s\033[0m", "33", s)
}
