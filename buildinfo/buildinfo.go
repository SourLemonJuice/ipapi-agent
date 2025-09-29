package buildinfo

import (
	"runtime/debug"
)

var (
	Version string
)

func init() {
	Version = "develop"
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
}
