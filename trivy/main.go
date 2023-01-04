package main

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/framework"
	"github.com/TencentBlueKing/ci-repoAnalysis/trivy/cmd"
)

func main() {
	framework.Analyze(cmd.TrivyExecutor{})
}
