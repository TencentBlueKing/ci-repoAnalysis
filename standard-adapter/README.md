## 标准扫描器适配器

对仅支持解析input.json文件扫描输出output.jso文件的离线扫描器进行增强，使其支持从服务端拉取任务执行并上报结果

```go
// 接入示例
func main() {
    // 将会从服务端拉取任务，将任务信息写入input.json后，
	// 在workdDir下执行/bin/scan --input /bkrepo/workspace/input.json --output /bkrepo/workspace/output.json
	framework.Analyze(StandardAdapterExecutor{cmd: "/bin/scan", workDir: "/bkrepo/workspace"})
}
```
