package useq

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer, err := New(Settings{})
	if err != nil {
		t.Fatal(err)
	}

	analysistest.Run(t, testdata, analyzer.Analyzer(), "test")
}
