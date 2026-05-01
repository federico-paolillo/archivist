package main

import (
	"context"
	"os"

	"codeberg.org/federico-paolillo/archivist/internal/app"
	"codeberg.org/federico-paolillo/archivist/internal/runner"
)

func main() {
	statusCode := runner.RunMany(
		context.Background(),
		app.CliProgram,
	)

	if statusCode == runner.NotOk {
		os.Exit(1)
	}
}
