package object

// ManifestV1 镜像manifest v1版本
type ManifestV1 struct {
	Config   string
	RepoTags []string
	Layers   []string
}
