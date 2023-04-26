package pkg

import (
	"testing"
)

func TestParsePackageNameAndVersion(t *testing.T) {
	pkgName, pkgVersion := ParsePackageNameAndVersion("npm-test-0.0.1.tgz")
	if pkgName != "npm-test" || pkgVersion != "0.0.1" {
		t.Fatalf("parese failed pkgName[%s] pkgVersion[%s]", pkgName, pkgVersion)
	}
}

func TestExtractPackageNameAndVersion(t *testing.T) {
	pkgName, pkgVersion, err := ExtractPackageNameAndVersion("testdata/axios-0.16.2.tgz")
	if err != nil {
		t.Fatal(err.Error())
	}
	if pkgName != "axios" || pkgVersion != "0.16.2" {
		t.Fatalf("parese failed pkgName[%s] pkgVersion[%s]", pkgName, pkgVersion)
	}
}
