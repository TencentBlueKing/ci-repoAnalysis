# 蓝盾制品库分析工具golang sdk
本SDK定义了分析工具的输入、输出，接入的扫描器仅需编写执行扫描器和输出output的逻辑，方便扫描器接入

## 快速上手
### 引入sdk
```gotemplate
    require github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang v0.0.9
```
### 使用示例
只需要实现Executor接口，再调用Analyze(executor)函数即可，参考下方示例
```gotemplate
// 1. 实现Executor接口
type SimpleExecutor struct {}

func (e *SimpleExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
    // 从config中获取配置	
    packageType := config.GetStringArg(util.ArgKeyPkgType)
    if packageType != util.PackageTypeDocker {
        return nil, errors.New("Package type [" + packageType + "] was not supported")
    }
    // 执行分析
    log.Println(file.Name())

    // 返回结果 
    return object.NewOutput("SUCCESS", &object.Result{}), nil
}

// 2. 执行分析
func main() {
    framework.Analyze(new(SimpleExecutor))
}
```
