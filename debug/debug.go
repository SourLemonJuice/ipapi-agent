package debug

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SourLemonJuice/ipapi-agent/build"
	"github.com/fatih/color"
)

var (
	// set the flag other than 0 may cause some performance issue
	Logger *log.Logger = log.New(io.Discard, "[debug] ", 0)
)

func Enable() {
	Logger.SetOutput(os.Stdout)
}

func PrintIntro() {
	fmt.Println("======Version======")
	build.PrintVersion()
	fmt.Println("======End Version======")
	Logger.Printf("Has NoColor flag: %v\n", color.NoColor)
	fmt.Println("======End Debug======")
}
