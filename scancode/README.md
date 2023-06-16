## Scancode制品分析工具

[scancode-toolkit](https://github.com/nexB/scancode-toolkit)扫描制品使用的License

### 使用方式

一、直接使用已经构建好的镜像

`ghcr.io/TencentBlueKing/ci-repoanalysis/bkrepo-scancode:latest`

---

二、手动构建镜像

1. 在scancode目录下执行命令构建分析工具镜像`go mod download && go build -o bkrepo-scancode ./cmd/main.go && docker build . -t bkrepo-scancode:0.0.1`
2. 将构建好的镜像推送到镜像仓库

---

最后在蓝鲸制品库Admin中配置`Standard`类型的扫描器，
启动命令为`/bkrepo-scancode`
