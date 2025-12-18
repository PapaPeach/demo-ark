//go:build !windows

package configio

import (
	"demo-ark/demoark/internal/util"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

/* Creates a shortcut to run the program with the currently applied options. */
func CreateConfiguredShortcut(args Arguments) {
	// Get filepath of program
	programPath, err := os.Executable()
	if err != nil {
		log.Println("Error getting program path for shortcut:", err)
		util.EnterToExit(false)
	}

	// Get currently applied options
	argsString := "Silent=1"
	argsTypes := reflect.TypeOf(args)
	argsValues := reflect.ValueOf(args)
	for i := 1; i < argsTypes.NumField(); i++ {
		// Format currently applied options as expected arguments
		value := 0
		switch argsValues.Field(i).Kind() {
		case reflect.Bool:
			// Skip create shortcut
			if strings.EqualFold(argsTypes.Field(i).Name, CreateShortcut) {
				continue
			}
			// Simplify booleans arguments
			if argsValues.Field(i).Bool() {
				value = 1
			}
			argsString += fmt.Sprintf(" %s=%d", argsTypes.Field(i).Name, value)

		case reflect.Uint8: // Int arguments
			fallthrough
		case reflect.Uint16:
			argsString += fmt.Sprintf(" %s=%v", argsTypes.Field(i).Name, argsValues.Field(i))

		case reflect.String: // String arguments
			if len(argsValues.Field(i).String()) == 0 {
				continue
			}
			argsString += fmt.Sprintf(" %s=%s", argsTypes.Field(i).Name, argsValues.Field(i))

		case reflect.Slice: // Slice argument
			if argsValues.Field(i).Len() == 0 {
				continue
			}
			argsString += fmt.Sprintf(" %s=", argsTypes.Field(i).Name)
			for j := range argsValues.Field(i).Len() {
				argsString += fmt.Sprintf("%s ", argsValues.Field(i).Index(j))
			}
		default:
			log.Println("Error getting arguments for shortcut", err)
			util.EnterToExit(false)
		}
	}

	// Get filepath to TF2's game.ico
	path, err := os.Getwd()
	if err != nil {
		log.Println("Error getting working directory for shortcut icon:", err)
	}

	// Make sure path is in tf directory
	tf := filepath.Join("Team Fortress 2", "tf")
	iconPath := path
	tfIndex := strings.Index(iconPath, tf)
	if tfIndex == -1 {
		fmt.Println("Program is not currently in a subdirectory of tf, cannot locate TF2 icon.")
	}
	iconPath = filepath.Join(iconPath[:tfIndex+len(tf)], "resource")

	// Get correct icon for os
	switch runtime.GOOS {
	case "linux":
		iconPath = filepath.Join(iconPath, "game.ico")
	case "darwin": // Mac
		iconPath = filepath.Join(iconPath, "game.icns")
	default:
		log.Println("Detected unsupported operating system.")
		util.EnterToExit(false)
	}

	// Create shortcut
	shortcutPath := filepath.Join(".", "Demo Ark")
	switch runtime.GOOS {
	case "linux": // Create .desktop shortcut
		fmt.Println("Creating configured Linux shortcut...")
		sc, err := os.Create(shortcutPath + ".desktop")
		if err != nil {
			log.Println("Error creating Linux shortcut:", err)
		}
		defer sc.Close()

		// Write contents of .desktop file
		sc.WriteString("[Desktop Entry]\n")
		sc.WriteString("Type=Application\n")
		sc.WriteString("Version=1.0\n")

		sc.WriteString("Name=Demo Ark\n")
		sc.WriteString("Comment=Demo archiver with user configured options\n")
		sc.WriteString("Icon=" + iconPath + "\n")
		sc.WriteString("Path=" + path + "\n")
		sc.WriteString("Exec=\"" + programPath + "\" " + argsString + "\n")

		sc.WriteString("Terminal=false\n")
		sc.WriteString("Categories=Games;Utility;Application;\n")

	case "darwin": // Mac: Just print path for copy + paste
		fmt.Println("Creating copy+paste-able Mac path...")
		fmt.Printf("%s %s\nPath to icon: %s\n", programPath, argsString, iconPath)
	default:
		log.Println("Detected unsupported operating system.")
		util.EnterToExit(false)
	}
}
