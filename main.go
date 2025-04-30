package main

import (
	"context"
	"fmt"

	"github.com/phuslu/log"

	"github.com/joho/godotenv"
	"github.com/prigas-dev/backoffice-ai/http_server"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("error loading .env file"))
	}

	ctx := context.Background()

	http_server.Start(ctx)
}
