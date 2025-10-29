package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

/** Scan demos and determine gamemode type */
func groupGameTypes(demos []string) ([]string, []string, []string) {
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
			// Check if tournament 1, stopwatch 0, readymode 1, readymode_min 0
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
			casualDemos = append(casualDemos, filename)
		} else {
			tournamentDemos = append(tournamentDemos, filename)
		}
		file.Close()
	}
	return casualDemos, mvmDemos, tournamentDemos
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
	demoCount := len(demos)
	fmt.Printf("Scanning %d demos...\n", demoCount)

	// Asynchronously scan demos if there's a lot
	var casualDemos []string
	var mvmDemos []string
	var tournamentDemos []string
	cores := (runtime.NumCPU() / 2) - 1
	if cores > 1 && demoCount > cores*10 {
		type GameTypes struct {
			casual     []string
			mvm        []string
			tournament []string
		}

		// Determine how to split demos list to distribute across threads
		chunkLength := demoCount / cores
		chunkStart := 0
		chunkEnd := chunkLength
		fmt.Printf("Utilizing %d cores...\n", cores)

		// Scan demos and determine gamemode type
		scanned := make([]GameTypes, cores)
		var scanners sync.WaitGroup
		for i := range cores - 1 {
			cs := chunkStart
			ce := chunkEnd
			scanners.Go(func() {
				c, m, t := groupGameTypes(demos[cs:ce])
				scanned[i] = GameTypes{c, m, t}
			})
			chunkStart += chunkLength
			chunkEnd += chunkLength
		}
		scanners.Go(func() {
			c, m, t := groupGameTypes(demos[chunkStart:])
			scanned[cores-1] = GameTypes{c, m, t}
		})
		scanners.Wait()

		// Combine scanned gametype demos
		for i := range scanned {
			casualDemos = append(casualDemos, scanned[i].casual...)
			mvmDemos = append(mvmDemos, scanned[i].mvm...)
			tournamentDemos = append(tournamentDemos, scanned[i].tournament...)
		}
	} else { // Scan using single thread
		casualDemos, mvmDemos, tournamentDemos = groupGameTypes(demos)
	}

	// Move files to corresponding directories, asynchronously!
	var movers sync.WaitGroup
	movers.Go(func() { moveToDirectory("demos_casual", casualDemos) })
	movers.Go(func() { moveToDirectory("demos_mvm", mvmDemos) })
	movers.Go(func() { moveToDirectory("demos_tournament", tournamentDemos) })
	movers.Wait()

	//enterToExit()
}
