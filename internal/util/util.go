package util

import (
	"fmt"
	"os"
)

/* Prompts user to press enter to exit the program, then exit when entered. */
func EnterToExit(silent bool) {
	if !silent {
		fmt.Println("Press enter to close.")
		fmt.Scanln()
	}
	os.Exit(0)
}
