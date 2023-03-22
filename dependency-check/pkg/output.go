package pkg

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"strings"
)

// Report Dependency-Check输出的报告
type Report struct {
	Dependencies []Dependency `json:"dependencies"`
}

// Dependency 检出的依赖及漏洞
type Dependency struct {
	FileName        string          `json:"fileName"` // 被扫描的的文件名
	FilePath        string          `json:"filePath"` // 在被扫描包中的路径
	Packages        []Package       `json:"packages"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

// Package 检出漏洞的包
type Package struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

// Vulnerability 漏洞信息
type Vulnerability struct {
	Name        string      `json:"name"`     // 漏洞ID
	Severity    string      `json:"severity"` // 漏洞威胁等级CRITICAL、HIGH、MEDIUM、LOW
	CCSSV2      CVSSV2      `json:"cvssv2"`
	CVSSV3      CVSSV3      `json:"cvssv3"`
	Description string      `json:"description"` // 漏洞描述
	References  []Reference `json:"references"`  // 漏洞相关的链接
}

// CVSSV2 cvssV2信息
type CVSSV2 struct {
	Score               float64 `json:"score"`
	AccessVector        string  `json:"accessVector"`
	AccessComplexity    string  `json:"accessComplexity"`
	Authenticationr     string  `json:"authenticationr"`
	ConfidentialImpact  string  `json:"confidentialImpact"`
	IntegrityImpact     string  `json:"integrityImpact"`
	AvailabilityImpact  string  `json:"availabilityImpact"`
	Severity            string  `json:"severity"`
	Version             string  `json:"version"`
	ExploitabilityScore string  `json:"exploitabilityScore"`
	ImpactScore         string  `json:"impactScore"`
}

// CVSSV3 cvssV3信息
type CVSSV3 struct {
	BaseScore             float64 `json:"baseScore"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
	BaseSeverity          string  `json:"baseSeverity"`
	ExploitabilityScore   string  `json:"exploitabilityScore"`
	ImpactScore           string  `json:"impactScore"`
	Version               string  `json:"version"`
}

// Reference 漏洞相关链接
type Reference struct {
	Url string `json:"url"`
}

// Convert 转换工具输出的报告为制品库需要的报告
func Convert(report *Report) *object.Result {
	result := new(object.Result)

	for i := range report.Dependencies {
		dependency := report.Dependencies[i]
		pkgName, pkgVersion := parsePkg(dependency.Packages)
		for k := range dependency.Vulnerabilities {
			vul := dependency.Vulnerabilities[k]
			securityResult := convertToSecurityResult(&vul, dependency.FilePath, pkgName, pkgVersion)
			result.SecurityResults = append(result.SecurityResults, *securityResult)
		}
	}

	return result
}

func convertToSecurityResult(
	vul *Vulnerability,
	filePath string,
	pkgName string,
	pkgVersion string,
) *object.SecurityResult {
	references := make([]string, len(vul.References))
	for i := range vul.References {
		references[i] = vul.References[i].Url
	}

	return &object.SecurityResult{
		VulId:           vul.Name,
		VulName:         vul.Name,
		CveId:           vul.Name,
		Path:            filePath,
		PkgName:         pkgName,
		PkgVersions:     []string{pkgVersion},
		EffectedVersion: "",
		FixedVersion:    "",
		Des:             vul.Description,
		Solution:        "",
		References:      references,
		Cvss:            vul.CVSSV3.BaseScore,
		Severity:        vul.Severity,
	}
}

func parsePkg(packages []Package) (string, string) {
	if len(packages) == 0 {
		return "", ""
	} else {
		pkg := packages[0]
		pkgParts := strings.Split(pkg.Id, "/")
		versionParts := strings.Split(pkgParts[len(pkgParts)-1], "@")
		version := versionParts[len(versionParts)-1]
		return strings.Join(pkgParts[1:len(pkgParts)-1], ":") + ":" + versionParts[0], version
	}
}
