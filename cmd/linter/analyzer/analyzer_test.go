package analyzer_test

import (
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "a", "mainpkg")
}
