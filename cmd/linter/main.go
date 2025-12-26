package main

import (
	"github.com/makimaki04/go-metrics-agent.git/internal/linter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(&linter.PanicCheckAnalyzer)
}
