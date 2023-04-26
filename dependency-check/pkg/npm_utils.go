package pkg

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"io"
	"os"
	"strings"
)

// ParsePackageNameAndVersion 从文件名解析包名和版本
func ParsePackageNameAndVersion(fileBaseName string) (string, string) {
	// 获取 pkgName 和 pkgVersion
	indexOfLastHyphens := strings.LastIndex(fileBaseName, "-")
	if indexOfLastHyphens == -1 {
		return "", ""
	}
	indexOfLastDot := strings.LastIndex(fileBaseName, ".")
	if indexOfLastDot == -1 {
		return "", ""
	}
	pkgName := fileBaseName[:indexOfLastHyphens]
	pkgVersion := fileBaseName[indexOfLastHyphens+1 : indexOfLastDot]
	util.Info("npm package %s, version %s", pkgName, pkgVersion)

	return pkgName, pkgVersion
}

// ExtractPackageNameAndVersion 从package.json文件中解析出packageName、version
func ExtractPackageNameAndVersion(npmPkgPath string) (string, string, error) {
	f, err := os.Open(npmPkgPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	uncompressedStream, err := gzip.NewReader(f)
	if err != nil {
		return "", "", err
	}
	defer uncompressedStream.Close()
	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}
		if header.Typeflag == tar.TypeReg && header.Name == "package/package.json" {
			npmPkg := &npmPackage{}
			if err := json.NewDecoder(tarReader).Decode(npmPkg); err != nil {
				return "", "", err
			}
			return npmPkg.Name, npmPkg.Version, nil
		}
	}

	return "", "", nil
}

type npmPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
