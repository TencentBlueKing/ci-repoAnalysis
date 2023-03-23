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

### 离线扫描

在无法访问外网的环境，可以在制品库Admin中为扫描器增加下面的参数

1. boolean类型参数`offline`设置为true
2. string类型参数`dbUrl`设置为漏洞库的下载链接

#### 漏洞库创建

在dependency-check镜像中执行`/usr/share/dependency-check/bin/dependency-check.sh --updateonly`后，
将`/usr/share/dependency-check/data`路径下的`odc.mv.db`、`publishedSuppressions.xml`、`jsrepository.json`打包成tar.gz
上传到执行扫描的环境可访问的位置即可


