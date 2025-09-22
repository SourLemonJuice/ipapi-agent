package buildinfo

import (
	"runtime"
	"runtime/debug"
)

var Version string
var GoVersion string
var OS string
var Arch string

func init() {
	Version = "develop"
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}

	GoVersion = runtime.Version()

	OS = runtime.GOOS
	Arch = runtime.GOARCH
}
