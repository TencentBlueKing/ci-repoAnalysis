## Trivy制品分析工具

[trivy](https://github.com/aquasecurity/trivy)分析工具可用于发现镜像中存在漏洞、敏感信息等问题

### 使用方式

1. 在trivy目录下执行命令构建分析工具镜像`go mod download && go build && docker build . -t bkrepo-trivy:0.0.1`
2. 将构建好的镜像推送到镜像仓库
3. 在蓝鲸制品库Admin中配置Standard类型的扫描器，并添加`STRING`类型的参数`dbDownloadUrl`指定漏洞库下载地址
