package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

/** Checks if conVar exists with the desired value in a array conVars */
func checkConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/** Returns a list of .dem files in the current directory */
func getDemos() []string {
	// Get list of files in current directory
	files, err := os.ReadDir(".")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Filter list to only have .dem files
	var demos []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dem") {
			demos = append(demos, file.Name())
		}
	}

	return demos
}

/** Moves files in a list of files to a directory */
func moveToDirectory(wishDir string, files []string) {
	// Handle length accordingly
	length := len(files)
	switch length {
	case 0: // Skip if no files to move exist
		return
	case 1: // Singular demo file
		fmt.Printf("Moving %d demo to %s...\n", length, wishDir)
	default: // Plural demos
		fmt.Printf("Moving %d demos to %s...\n", length, wishDir)
	}

	// Make directory to move to
	err := os.Mkdir(wishDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Println(err)
		os.Exit(1)
	}

	// Move files to directory
	for _, file := range files {
		err := os.Rename(file, filepath.Join(wishDir, file))
		if err != nil {
			log.Println(err)
		}
	}
}

/** Prompts user to press enter to exit the program, then exit when entered */
func enterToExit() {
	fmt.Println("Press enter to close.")
	fmt.Scanln()
	os.Exit(0)
}

func main() {
	// Get demos
	demos := getDemos()
	fmt.Printf("Scanning %d demos...\n", len(demos))

	// Scan demos and determine gamemode type
	var casualDemos []string
	var mvmDemos []string
	var tournamentDemos []string
	for _, filename := range demos {
		// Open file for reading
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening %v: %v", filename, err)
		}

		// Determine if MvM via map prefix in header
		header := readHeader(file)
		if strings.HasPrefix(header.MapName, "mvm_") {
			//fmt.Println("Detected mvm demo:\t\t", filename)
			mvmDemos = append(mvmDemos, filename)
			file.Close()
			continue
		}

		// Get message contents
		file.Seek(1072, io.SeekStart)
		msg := readMessage(file)

		// Determine if casual via specific conVar values
		casual := false
		for _, sc := range msg.ParsedData.SetConVar {
			// Check if tournament 1, stopwatch 0, tournament_readymode_min 0
			if checkConVar(sc.ConVars, "mp_tournament", "1") &&
				checkConVar(sc.ConVars, "mp_tournament_stopwatch", "0") &&
				checkConVar(sc.ConVars, "mp_tournament_readymode", "1") &&
				checkConVar(sc.ConVars, "mp_tournament_readymode_min", "0") {
				casual = true
				file.Close()
				break
			}
		}
		if casual {
			//fmt.Println("Detected casual demo:\t\t", filename)
			casualDemos = append(casualDemos, filename)
		} else {
			//fmt.Println("Detected tournament demo:\t", filename)
			tournamentDemos = append(tournamentDemos, filename)
		}
		file.Close()
	}

	// Move files to corresponding directories, asynchronously!
	var wg sync.WaitGroup
	wg.Go(func() { moveToDirectory("demos_casual", casualDemos) })
	wg.Go(func() { moveToDirectory("demos_mvm", mvmDemos) })
	wg.Go(func() { moveToDirectory("demos_tournament", tournamentDemos) })
	wg.Wait()

	//enterToExit()
}
