package main

import (
	"fmt"
	"os"

	"github.com/dhaus67/useq"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	analyzer, err := useq.NewAnalyzer(useq.Settings{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	singlechecker.Main(analyzer)
}
