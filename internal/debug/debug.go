package debug

import (
	"fmt"

	"github.com/fatih/color"
)

func Print() {
	fmt.Printf("NoColor flag: %v\n", color.NoColor)
}
