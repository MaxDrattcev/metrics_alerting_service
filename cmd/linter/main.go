package main

import (
	"github.com/MaxDrattcev/metrics_alerting_service/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
