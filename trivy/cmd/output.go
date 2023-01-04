package cmd

import (
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/trivy/constant"
	"strings"
	"time"
)

// Output trivy扫描输出
type Output struct {
	SchemaVersion int      `json:"SchemaVersion"`
	ArtifactName  string   `json:"ArtifactName"`
	ArtifactType  string   `json:"ArtifactType"`
	Results       []Result `json:"Results"`
}

// Result trivy扫描结果
type Result struct {
	Target          string          `json:"Target"`
	Class           string          `json:"Class"`
	Type            string          `json:"Type"`
	Vulnerabilities []Vulnerability `json:"Vulnerabilities"`
	Secrets         []Secret        `json:"Secrets"`
}

// Vulnerability trivy扫描漏洞
type Vulnerability struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	PkgPath          string `json:"PkgPath"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Layer            struct {
		DiffID string `json:"DiffID"`
	} `json:"Layer"`
	SeveritySource string `json:"SeveritySource"`
	PrimaryURL     string `json:"PrimaryURL"`
	DataSource     struct {
		ID   string `json:"ID"`
		Name string `json:"Name"`
		URL  string `json:"URL"`
	} `json:"DataSource"`
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Severity    string   `json:"Severity"`
	CweIDs      []string `json:"CweIDs"`
	CVSS        struct {
		Ghsa struct {
			V3Vector string  `json:"V3Vector"`
			V3Score  float64 `json:"V3Score"`
		} `json:"ghsa"`
		Nvd struct {
			V2Vector string  `json:"V2Vector"`
			V3Vector string  `json:"V3Vector"`
			V2Score  float64 `json:"V2Score"`
			V3Score  float64 `json:"V3Score"`
		} `json:"nvd"`
		Redhat struct {
			V3Vector string  `json:"V3Vector"`
			V3Score  float64 `json:"V3Score"`
		} `json:"redhat"`
	} `json:"CVSS"`
	References       []string  `json:"References"`
	PublishedDate    time.Time `json:"PublishedDate"`
	LastModifiedDate time.Time `json:"LastModifiedDate"`
}

// Secret trivy敏感信息扫描结果
type Secret struct {
	RuleID    string `json:"RuleID"`
	Category  string `json:"Category"`
	Severity  string `json:"Severity"`
	Title     string `json:"Title"`
	StartLine int    `json:"StartLine"`
	EndLine   int    `json:"EndLine"`
	Match     string `json:"Match"`
}

// ConvertToToolResults 转换trivy扫描结果为工具框架标准扫描结果
func ConvertToToolResults(output *Output) *object.Result {
	toolResults := new(object.Result)

	if output.Results == nil {
		return toolResults
	}

	for _, result := range output.Results {
		if result.Class == constant.ClassSecret {
			for _, secret := range result.Secrets {
				toolResults.SensitiveResults =
					append(toolResults.SensitiveResults, *ConvertToSensitiveResult(&secret, result.Target))
			}
		} else {
			for _, vulnerability := range result.Vulnerabilities {
				toolResults.SecurityResults =
					append(toolResults.SecurityResults, *ConvertToSecurityResult(&vulnerability))
			}
		}
	}
	return toolResults
}

// ConvertToSecurityResult 转换trivy漏洞扫描结果为工具框架漏洞扫描结果
func ConvertToSecurityResult(vulnerability *Vulnerability) *object.SecurityResult {
	var references []string
	if vulnerability.References == nil {
		references = []string{}
	} else {
		references = vulnerability.References
	}
	return &object.SecurityResult{
		VulId:           vulnerability.VulnerabilityID,
		VulName:         vulnerability.Title,
		CveId:           vulnerability.VulnerabilityID,
		Path:            vulnerability.PkgPath,
		PkgName:         vulnerability.PkgName,
		PkgVersions:     []string{vulnerability.InstalledVersion},
		EffectedVersion: vulnerability.InstalledVersion,
		FixedVersion:    vulnerability.FixedVersion,
		Des:             vulnerability.Description,
		Solution:        "",
		References:      references,
		Cvss:            vulnerability.CVSS.Nvd.V3Score,
		Severity:        strings.ToLower(vulnerability.Severity),
	}
}

// ConvertToSensitiveResult 转换trivy敏感信息扫描结果为工具框架敏感信息扫描结果
func ConvertToSensitiveResult(secret *Secret, path string) *object.SensitiveResult {
	return &object.SensitiveResult{
		Path:    path,
		Type:    secret.Category,
		Content: secret.Match,
	}
}
