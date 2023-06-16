package pkg

import (
	"fmt"
	"testing"
)

func TestTransform(t *testing.T) {
	result, _ := transform("testdata/result.json")
	fmt.Println(result.Result.LicenseResults[0].Path)
	fmt.Println(result.Result.LicenseResults[0].LicenseName)
}
