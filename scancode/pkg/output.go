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
	Path                    string   `json:"path"`
	Type                    string   `json:"type"`
	SpdxLicenseKey          string   `json:"detected_license_expression_spdx"`
	LicenseExpressions      []string `json:"license_expressions"`
	PercentageOfLicenseText float64  `json:"percentage_of_license_text"`
	ScanErrors              []string `json:"scan_errors"`
}

// Convert 转换工具输出的报告为制品库需要的报告
func Convert(report *Report) *object.Result {
	result := new(object.Result)
	licenseResults := make(map[object.LicenseResult]struct{})
	var Empty struct{}

	for i := range report.Files {
		file := report.Files[i]
		path := file.Path[strings.Index(file.Path, "/")+1:]
		licenses := parseLicense(file.SpdxLicenseKey)
		for j := range licenses {
			license := licenses[j]
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

func parseLicense(detectedLicense string) []string {
	splitLicenses := strings.Split(detectedLicense, " ")
	licenses := make([]string, 0)
	for i := range splitLicenses {
		l := splitLicenses[i]
		if l == "AND" || l == "OR" {
			continue
		} else if l[0] == '(' {
			licenses = append(licenses, l[1:])
		} else if l[len(l)-1] == ')' {
			licenses = append(licenses, l[0:len(l)-1])
		} else {
			licenses = append(licenses, l)
		}
	}
	return licenses
}
