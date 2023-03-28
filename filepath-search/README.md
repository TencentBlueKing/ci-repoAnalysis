# FilepathSearch
在镜像Tar包中搜索匹配指定正则表达式的路径

## 使用方式
在制品库Admin中添加standalone类型的分析工具，并设置下列参数

1. 镜像地址`ghcr.io/TencentBlueKing/ci-repoanalysis/bkrepo-filepath-search:latest`
2. 增加string类型的参数regex，值为用于路径匹配的正则表达式
