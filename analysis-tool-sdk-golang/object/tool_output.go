package object

type TaskStatus string

const (
	// StatusFailed 失败
	StatusFailed TaskStatus = "FAILED"

	// StatusSuccess 成功
	StatusSuccess TaskStatus = "SUCCESS"

	// StatusTimeout 超时
	StatusTimeout TaskStatus = "TIMEOUT"

	// StatusStopped 被中止
	StatusStopped TaskStatus = "STOPPED"
)

// ToolOutput 工具输出
type ToolOutput struct {
	Status TaskStatus `json:"status"`
	Err    string     `json:"err"`
	TaskId string     `json:"taskId"`
	Result *Result    `json:"result"`
}

// Result 工具扫描结果
type Result struct {
	SecurityResults  []SecurityResult  `json:"securityResults"`
	LicenseResults   []LicenseResult   `json:"licenseResults"`
	SensitiveResults []SensitiveResult `json:"sensitiveResults"`
}

// SecurityResult 工具扫描安全结果
type SecurityResult struct {
	VulId           string   `json:"vulId"`
	VulName         string   `json:"vulName"`
	CveId           string   `json:"cveId"`
	Path            string   `json:"path"`
	PkgName         string   `json:"pkgName"`
	PkgVersions     []string `json:"pkgVersions"`
	EffectedVersion string   `json:"effectedVersion"`
	FixedVersion    string   `json:"fixedVersion"`
	Des             string   `json:"des"`
	Solution        string   `json:"solution"`
	References      []string `json:"references"`
	Cvss            float64  `json:"cvss"`
	Severity        string   `json:"severity"`
}

// LicenseResult 工具扫描许可证结果
type LicenseResult struct {
	LicenseName string `json:"licenseName"`
	Path        string `json:"path"`
	PkgName     string `json:"pkgName"`
	PkgVersion  string `json:"pkgVersion"`
}

// SensitiveResult 工具扫描敏感信息结果
type SensitiveResult struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// NewErrorOutput 创建错误输出
func NewErrorOutput(err error, status TaskStatus) *ToolOutput {
	return &ToolOutput{
		Status: status,
		Err:    err.Error(),
	}
}

// NewFailedOutput 创建错误输出
func NewFailedOutput(err error) *ToolOutput {
	return &ToolOutput{
		Status: StatusFailed,
		Err:    err.Error(),
	}
}

// NewOutput 创建工具标准输出
func NewOutput(status TaskStatus, result *Result) *ToolOutput {
	return &ToolOutput{
		Status: status,
		Result: result,
	}
}
