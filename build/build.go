package build

import (
	"fmt"
	"runtime"
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

func PrintVersion() {
	fmt.Printf("ipapi-agent version %v\n\n", Version)

	fmt.Printf("Environment: %v %v/%v\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
