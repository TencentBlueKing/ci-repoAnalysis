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

// FlagSkipJavaDbUpdate 跳过java漏洞库更新
const FlagSkipJavaDbUpdate = "--skip-java-db-update"

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
const DbCacheDir = "/root/.cache/trivy"

// DbDir 漏洞库存放目录
const DbDir = "db"

// JavaDbDir java漏洞库存放目录
const JavaDbDir = "java-db"

// FlagSecretConfig 敏感信息规则配置文件
const FlagSecretConfig = "--secret-config"

// SecretRuleFilePath 敏感信息规则文件路径
const SecretRuleFilePath = "/rule.yaml"

// FormatJson json输出格式
const FormatJson = "json"

// OutputPath 输出文件路径
const OutputPath = util.WorkDir + "/trivy-output.json"

// CheckVuln 检查安全漏洞
const CheckVuln = "vuln"

const ClassSecret = "secret"
