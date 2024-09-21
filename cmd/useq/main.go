package main

import (
	"fmt"
	"os"

	"github.com/dhaus67/useq"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	useqAnalyzer, err := useq.New(useq.Settings{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	singlechecker.Main(useqAnalyzer.Analyzer())
}
