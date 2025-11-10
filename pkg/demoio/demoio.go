package demoio

import (
	"demo-ark/demoark/pkg/parser"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Demo struct {
	Name     string
	NewName  string
	Map      string
	DateTime time.Time
	Duration float32
}

/* Culls demos shorter than a specified minimum length */
func CullShortDemos(demos *[]Demo, min uint8) []Demo {
	var cull []Demo
	keep := (*demos)[:0]
	for _, demo := range *demos {
		if demo.Duration < float32(min) { // If short than minimum, move it to cull list
			cull = append(cull, demo)
		} else { // If longer than minimum, keep it in main demo list
			keep = append(keep, demo)
		}
	}

	// Update demos list
	*demos = keep
	return cull
}

/* Generates a time format string based on arguments */
func GetTimeFormat(twelveHourTime bool) string {
	// Format time
	if twelveHourTime {
		return "03-04-05"
	}
	return "15-04-05"
}

/* Gets the date from a demo's name */
func GetDateTime(demo Demo) time.Time {
	// Search for valid date my locating year via: prefix[20]YY-MM-DD_HH-MM-SS
	skipped := 0
	for i := strings.Index(demo.Name, "20"); i > -1; i = strings.Index(demo.Name[skipped:], "20") {
		dateIndex := skipped + i
		skipped += i + 1

		titleDateTime, err := time.Parse("2006-01-02_15-04-05", demo.Name[dateIndex:dateIndex+19])
		if err != nil { // This might error. We keep searching or fallback to demo's ModTime
			fmt.Println(err)
			continue
		}

		return titleDateTime
	}

	// Fallback value
	return demo.DateTime
}

/* Generates new names for demos to according to the arguments provided */
func GetNewName(demo Demo, dateTimeFormat string, keepPrefix bool, renameMap bool, renameDuration bool) {
	// Get prefix
	var wishName string
	if keepPrefix {
		// Locate prefix by indexing off year
		yearString := strconv.Itoa(demo.DateTime.Year())
		if yearIdx := strings.Index(demo.Name, yearString); yearIdx != -1 {
			wishName = demo.Name[:yearIdx] + "_"
		} else {
			demIdx := strings.Index(demo.Name, ".dem")
			wishName = demo.Name[:demIdx] + "_"
		}
	}

	// Add map name
	if renameMap {
		wishName += demo.Map + "_"
	}

	// Add date and time of creation / edit
	wishName += demo.DateTime.Format(dateTimeFormat)

	// Add duration
	if renameDuration {
		length := int(demo.Duration)
		wishName = fmt.Sprintf("%s_%d-%02d", wishName, length/60, length%60)
	}

	// Add .dem and set demo's new name
	demo.NewName = wishName + ".dem"
}

/* Returns a list of .dem files in the current directory */
func GetDemos(ignoreWords []string) []Demo {
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
			// Skip ignored words
			for _, ignoreWord := range ignoreWords {
				if strings.Contains(strings.ToLower(file.Name()), ignoreWord) {
					goto ignored
				}
			}

			// Get header for map name and duration
			f, err := os.Open(file.Name())
			if err != nil {
				fmt.Printf("Error opening %v: %v", file.Name(), err)
			}
			header := parser.ReadHeader(f)

			demo := Demo{Name: file.Name(), NewName: file.Name(), Map: header.MapName, DateTime: file.ModTime(), Duration: header.PlaybackTime}
			demos = append(demos, demo)
			f.Close()
		}
	ignored:
	}

	return demos
}

/* Checks if conVar exists with the desired value in a array conVars */
func CheckConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/* Group demos by gametype */
func GroupGameTypes(demos []Demo) ([]Demo, []Demo, []Demo) {
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
		msg := parser.ReadMessage(file)

		// Determine if casual via specific conVar values
		casual := false
		for _, sc := range msg.ParsedData.SetConVar {
			// Check if tournament 1, stopwatch 0, readymode 1, readymode_min 0
			if CheckConVar(sc.ConVars, "mp_tournament", "1") &&
				CheckConVar(sc.ConVars, "mp_tournament_stopwatch", "0") &&
				CheckConVar(sc.ConVars, "mp_tournament_readymode", "1") &&
				CheckConVar(sc.ConVars, "mp_tournament_readymode_min", "0") {
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
func GroupYears(demos []Demo) map[int][]Demo {
	years := make(map[int][]Demo)
	for _, demo := range demos {
		years[demo.DateTime.Year()] = append(years[demo.DateTime.Year()], demo)
	}

	return years
}

// TODO: Update _events.json
/* Moves files in a list of files to a directory */
func MoveToDirectory(wishDir string, demos []Demo) {
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
		err := os.Rename(demo.Name, filepath.Join(wishDir, demo.NewName))
		if err != nil {
			log.Println(err)
		}
	}
}
