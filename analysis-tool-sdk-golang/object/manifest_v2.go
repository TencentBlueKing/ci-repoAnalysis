package object

import "strings"

// ManifestV2 镜像manifest v2版本
type ManifestV2 struct {
	SchemaVersion int
	MediaType     string
	Config        Layer
	Layers        []Layer
}

// Layer 镜像层信息
type Layer struct {
	MediaType string
	Size      int64
	Digest    string
}

// LayerCount 统计Layer数量，可能存在重复layer
func (m *ManifestV2) LayerCount() map[string]int {
	countMap := make(map[string]int, len(m.Layers))
	for i := range m.Layers {
		l := m.Layers[i]
		s := l.Sha256()
		c := 0
		if _, ok := countMap[s]; ok {
			c = countMap[s]
		}
		countMap[s] = c + 1
	}
	return countMap
}

// Sha256 镜像层sha256
func (l *Layer) Sha256() string {
	digestSplits := strings.Split(l.Digest, ":")
	if len(digestSplits) != 2 || digestSplits[0] != "sha256" {
		panic("layer digest[" + l.Digest + "] is illegal")
	}
	return digestSplits[1]
}

// Filename 镜像层的文件名
func (l *Layer) Filename() string {
	return strings.Replace(l.Digest, ":", "__", 1)
}
