## Trivy制品分析工具

[trivy](https://github.com/aquasecurity/trivy)分析工具可用于发现镜像中存在漏洞、敏感信息等问题

### 使用方式

一、直接使用已经构建好的镜像

`ghcr.io/TencentBlueKing/ci-repoanalysis/bkrepo-trivy:latest`

---

二、手动构建镜像

1. 在trivy目录下执行命令构建分析工具镜像`go mod download && go build -o bkrepo-trivy && docker build . -t bkrepo-trivy:0.0.1`
2. 将构建好的镜像推送到镜像仓库

---

最后在蓝鲸制品库Admin中配置`Standard`类型的扫描器，
启动命令为`/bkrepo-trivy`，参考trivy Air-Gapped Environment配置文档下载`db.tar.gz`和`javadb.tar.gz`传到制品分析服务可访问的路径，比如某个制品仓库中，并添加`STRING`类型的参数`dbDownloadUrl`与`javaDbDownloadUrl`指定trivy漏洞库`db.tar.gz`与`javadb.tar.gz`的下载地址
