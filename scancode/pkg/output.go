package pkg

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"strings"
)

// Report Scancode扫描报告
type Report struct {
	Files []Item `json:"files"`
}

// Item Scancode扫描结果项
type Item struct {
	Path                    string    `json:"path"`
	Type                    string    `json:"type"`
	Licenses                []SubItem `json:"licenses"`
	LicenseExpressions      []string  `json:"license_expressions"`
	PercentageOfLicenseText float64   `json:"percentage_of_license_text"`
	ScanErrors              []string  `json:"scan_errors"`
}

// SubItem 扫描到的License
type SubItem struct {
	SpdxLicenseKey string `json:"spdx_license_key"`
}

// Convert 转换工具输出的报告为制品库需要的报告
func Convert(report *Report) *object.Result {
	result := new(object.Result)
	licenseResults := make(map[object.LicenseResult]struct{})
	var Empty struct{}

	for i := range report.Files {
		file := report.Files[i]
		for j := range file.Licenses {
			license := file.Licenses[j].SpdxLicenseKey
			path := file.Path[strings.Index(file.Path, "/")+1:]
			licenseResult := object.LicenseResult{
				LicenseName: license,
				Path:        path,
			}
			licenseResults[licenseResult] = Empty
		}
	}

	for k := range licenseResults {
		result.LicenseResults = append(result.LicenseResults, k)
	}

	return result
}
