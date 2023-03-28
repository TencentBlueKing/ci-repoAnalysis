package main

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/framework"
	"github.com/TencentBlueKing/ci-repoAnalysis/filepath-search/pkg"
)

func main() {
	framework.Analyze(pkg.FilepathSearch{})
}
