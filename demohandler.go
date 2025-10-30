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
)

type Demo struct {
	Name     string
	Year     int
	Duration float32
	Map      string
}

/* Returns a list of .dem files in the current directory */
func getDemos() []Demo {
	// Get list of files in current directory
	directory, err := os.Open(".")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	files, err := directory.Readdir(0)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer directory.Close()

	// Filter list to only have .dem files
	var demos []Demo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dem") {
			// Get header for map name and duration
			f, err := os.Open(file.Name())
			if err != nil {
				fmt.Printf("Error opening %v: %v", file.Name(), err)
			}
			header := readHeader(f)

			demo := Demo{file.Name(), file.ModTime().Year(), header.PlaybackTime, header.MapName}
			demos = append(demos, demo)
			f.Close()
		}
	}

	return demos
}

/* Checks if conVar exists with the desired value in a array conVars */
func checkConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/* Group demos by gametype */
func groupGameTypes(demos []Demo) ([]Demo, []Demo, []Demo) {
	var casualDemos []Demo
	var mvmDemos []Demo
	var tournamentDemos []Demo
	for i := range demos {
		// Open file for reading
		filename := demos[i].Name

		// Determine if MvM via map prefix in header
		if strings.HasPrefix(demos[i].Map, "mvm_") {
			mvmDemos = append(mvmDemos, demos[i])
			continue
		}

		// Get message contents
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening %v: %v", filename, err)
		}
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
			casualDemos = append(casualDemos, demos[i])
		} else {
			tournamentDemos = append(tournamentDemos, demos[i])
		}
		file.Close()
	}
	return casualDemos, mvmDemos, tournamentDemos
}

/* Groups demos by year */
func groupYears(demos []Demo) map[int][]Demo {
	years := make(map[int][]Demo)
	for _, demo := range demos {
		years[demo.Year] = append(years[demo.Year], demo)
	}

	return years
}

/* Moves files in a list of files to a directory */
func moveToDirectory(wishDir string, demos []Demo) {
	// Handle length accordingly
	length := len(demos)
	switch length {
	case 0: // Skip if no files to move exist
		return
	case 1: // Singular demo file
		fmt.Printf("Moving %d demo to %s...\n", length, wishDir)
	default: // Plural demos
		fmt.Printf("Moving %d demos to %s...\n", length, wishDir)
	}

	// Make directory to move to
	err := os.MkdirAll(wishDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Println(err)
		os.Exit(1)
	}

	// Move files to directory
	for _, demo := range demos {
		err := os.Rename(demo.Name, filepath.Join(wishDir, demo.Name))
		if err != nil {
			log.Println(err)
		}
	}
}
