package useq

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer, err := NewAnalyzer(Settings{
		Functions: []string{"test.myCustomPrint"},
	})
	if err != nil {
		t.Fatal(err)
	}

	analysistest.RunWithSuggestedFixes(t, testdata, analyzer, "test")
}
