package debug

import (
	"fmt"

	"github.com/SourLemonJuice/ipapi-agent/internal/build"
	"github.com/fatih/color"
)

func Print() {
	fmt.Println("======Version======")
	build.PrintVersion()
	fmt.Println("======End Version======")
	fmt.Printf("has NoColor flag: %v\n", color.NoColor)
	fmt.Println("======End Debug Info======")
}
