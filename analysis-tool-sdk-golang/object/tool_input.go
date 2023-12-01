package object

import (
	"strconv"
	"time"
)

// ToolInput 工具输入
type ToolInput struct {
	TaskId     string     `json:"taskId"`
	ToolConfig ToolConfig `json:"toolConfig"`
	FilePath   string     `json:"filePath"`
	Sha256     string     `json:"sha256"`
	FileUrls   []FileUrl  `json:"fileUrls"`
}

// ToolConfig 工具配置
type ToolConfig struct {
	Args []Argument `json:"args"`
}

// Argument 工具配置参数
type Argument struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Des   string `json:"des"`
}

// FileUrl 文件下载地址
type FileUrl struct {
	Url    string `json:"url"`
	Name   string `json:"name"`
	Sha256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// GetBoolArg 获取布尔类型参数
func (toolConfig *ToolConfig) GetBoolArg(key string) (bool, error) {
	var argument Argument
	for _, arg := range toolConfig.Args {
		if arg.Key == key && arg.Type == "BOOLEAN" {
			argument = arg
			break
		}
	}
	return strconv.ParseBool(argument.Value)
}

// GetFloatArg 获取浮点类型参数
func (toolConfig *ToolConfig) GetFloatArg(key string) (float64, error) {
	var argument Argument
	for _, arg := range toolConfig.Args {
		if arg.Key == key && arg.Type == "NUMBER" {
			argument = arg
			break
		}
	}
	return strconv.ParseFloat(argument.Value, 64)
}

// GetIntArg 获取整形参数
func (toolConfig *ToolConfig) GetIntArg(key string) (int64, error) {
	var argument Argument
	for _, arg := range toolConfig.Args {
		if arg.Key == key && arg.Type == "NUMBER" {
			argument = arg
			break
		}
	}
	return strconv.ParseInt(argument.Value, 10, 64)
}

// GetStringArg 获取字符串类型参数
func (toolConfig *ToolConfig) GetStringArg(key string) string {
	var argument Argument
	for _, arg := range toolConfig.Args {
		if arg.Key == key && arg.Type == "STRING" {
			argument = arg
			break
		}
	}
	return argument.Value
}

// FileUrlMap 获取文件sh256到url的映射
func (toolInput *ToolInput) FileUrlMap() map[string]FileUrl {
	fileUrlMap := make(map[string]FileUrl, len(toolInput.FileUrls))
	for _, url := range toolInput.FileUrls {
		fileUrlMap[url.Sha256] = url
	}
	return fileUrlMap
}

// MaxTime 获取允许执行的最长时间
func (toolInput *ToolInput) MaxTime() time.Duration {
	maxTime, _ := toolInput.ToolConfig.GetIntArg("maxTime")
	return time.Duration(maxTime) * time.Millisecond
}
