package util

import (
	"fmt"
	"os"
)

/* Prompts user to press enter to exit the program, then exit when entered. */
func EnterToExit(silent bool, err error) {
	var exitCode int

	// Print error
	if err != nil {
		fmt.Println(err)
		exitCode = 1
	}

	// Keep window open till user presses enter
	if !silent || err != nil {
		_, _ = fmt.Println("Press enter to close.")
		_, _ = fmt.Scanln()
	}

	os.Exit(exitCode)
}
