package constant

import "github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"

// CmdTrivy trivy命令
const CmdTrivy = "trivy"

// SubCmdImage trivy image子命令
const SubCmdImage = "image"

// FlagInput 指定扫描文件
const FlagInput = "--input"

// FlagCacheDir 指定trivy缓存目录
const FlagCacheDir = "--cache-dir"

// FlagSecurityChecks 指定扫描类型
const FlagSecurityChecks = "--security-checks"

// FlagSkipDbUpdate 跳过更新trivy.db
const FlagSkipDbUpdate = "--skip-db-update"

// FlagOfflineScan 离线扫描
const FlagOfflineScan = "--offline-scan"

// FlagTimeout 超时时间
const FlagTimeout = "--timeout"

// FlagFormat 指定输出格式
const FlagFormat = "-f"

// FlagOutput 指定输出路径
const FlagOutput = "-o"

// CacheDir trivy缓存目录
const CacheDir = "/root/.cache/trivy"

// DbCacheDir trivy.db缓存目录
const DbCacheDir = "/root/.cache/trivy/db"

// DbFilename trivy.db文件名
const DbFilename = "trivy.db"

// FormatJson json输出格式
const FormatJson = "json"

// OutputPath 输出文件路径
const OutputPath = util.WorkDir + "/trivy-output.json"

// CheckVuln 检查安全漏洞
const CheckVuln = "vuln"

const ClassSecret = "secret"
