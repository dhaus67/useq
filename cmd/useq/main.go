package main

import (
	"fmt"
	"os"

	"github.com/dhaus67/useq"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	settings := useq.Settings{
		Validate: useq.DefaultValidationSettings,
	}

	useqAnalyzer, err := useq.New(settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	analyzers, err := useqAnalyzer.BuildAnalyzers()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	singlechecker.Main(analyzers[0])
}
