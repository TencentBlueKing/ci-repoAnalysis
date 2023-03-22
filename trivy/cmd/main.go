package main

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/framework"
	"github.com/TencentBlueKing/ci-repoAnalysis/trivy/pkg"
)

func main() {
	framework.Analyze(pkg.TrivyExecutor{})
}
