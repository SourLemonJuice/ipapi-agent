package debug

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SourLemonJuice/ipapi-agent/internal/build"
	"github.com/fatih/color"
)

var (
	// set the flag other than 0 may cause some performance issue
	Logger *log.Logger = log.New(io.Discard, "[DEBUG] ", 0)
)

func Enable() {
	Logger.SetOutput(os.Stdout)
}

func PrintIntro() {
	fmt.Println("======Version======")
	build.PrintVersion()
	fmt.Println("======End Version======")
	Logger.Printf("has NoColor flag: %v\n", color.NoColor)
}
