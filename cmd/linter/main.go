package main

import (
	"github.com/scarypuppp/metrics-service/cmd/linter/analyzer"
	_ "golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
