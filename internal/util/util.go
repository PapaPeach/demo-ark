package util

import (
	"fmt"
	"os"
)

/* Prompts user to press enter to exit the program, then exit when entered. */
func EnterToExit(silent bool) {
	if !silent {
		_, _ = fmt.Println("Press enter to close.")
		_, _ = fmt.Scanln()
	}
	os.Exit(0)
}
