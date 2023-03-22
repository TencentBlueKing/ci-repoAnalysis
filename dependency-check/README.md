## DependencyCheck制品分析工具

[DependencyCheck](https://github.com/jeremylong/DependencyCheck)分析工具可用于扫描制品中存在的漏洞

### 使用方式

一、直接使用已经构建好的镜像

`ghcr.io/TencentBlueKing/ci-repoanalysis/bkrepo-dependency-check:latest`

---

二、手动构建镜像
1. 设置环境变量`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
2. 拉取代码并执行`go mod download && go build -o bkrepo-dependency-check  ./cmd/main.go && docker build . -t bkrepo-dependency-check:latest`
3. 将构建好的镜像推送到镜像仓库

---

最后在蓝鲸制品库Admin中配置`Standard`类型的扫描器，启动命令设置为`/bkrepo-dependency-check`
