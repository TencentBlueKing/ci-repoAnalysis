package pkg

import (
	"fmt"
	"testing"
)

func TestReadToolOutput(t *testing.T) {
	output, err := StandardAdapterExecutor{}.readToolOutput("testdata/output.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(output.Result.SecurityResults)
}
