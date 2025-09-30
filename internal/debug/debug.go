package debug

import (
	"fmt"

	"github.com/fatih/color"
)

func Print() {
	fmt.Printf("has NoColor flag: %v\n", color.NoColor)
}
