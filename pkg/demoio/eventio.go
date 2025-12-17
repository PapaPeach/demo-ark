package demoio

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

/* Get _event.txt files */
func GetEventTxts(searchDirs bool, ignoreWords []string) (map[string][]string, []string) {
	// Filter list to only have _events.txt files
	eventTxts := make(map[string][]string)
	var culledEventTxts []string
	processFile := func(path string, file fs.DirEntry) {
		// Skip irrelevant files
		if file.IsDir() || !strings.HasSuffix(path, "_events.txt") {
			return
		}

		// If file is empty, skip it
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Printf("Error getting stats on %v: %v\n", path, err)
		}
		if fileInfo.Size() == 0 {
			culledEventTxts = append(culledEventTxts, path)
			return
		}

		// Get contents of _events.txt
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening %v to get contents: %v\n", path, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			// Isolate demo name
			var demoName string
			demoIndex := strings.Index(scanner.Text(), "(\"")
			if demoIndex != -1 {
				demoIndex += 2 // Trim (" from front
				endIndex := strings.Index(scanner.Text()[demoIndex:], "\"")
				if endIndex != -1 {
					demoName = scanner.Text()[demoIndex : demoIndex+endIndex-1]
				} else {
					continue
				}
			} else {
				continue
			}

			// If demo name is unique then add it to eventTxt map entry
			demoPath := strings.TrimSuffix(path, file.Name()) + demoName
			if !slices.Contains(eventTxts[path], demoPath) {
				eventTxts[path] = append(eventTxts[path], demoPath)
			}
		}
	}

	if searchDirs { // Search subdirectories
		filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error reading %v: %v\n", path, err)
				return nil
			}

			// Skip directories containing ignored words
			if file.IsDir() {
				for _, ignoreWord := range ignoreWords {
					if strings.Contains(strings.ToLower(file.Name()), ignoreWord) {
						return fs.SkipDir
					}
				}
			}

			processFile(path, file)
			return nil
		})
	} else { // Just search current directory
		// Get list of files in current directory
		dir, err := os.Open(".")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		files, err := dir.ReadDir(0)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer dir.Close()

		// Filter list to contain only demos
		for _, file := range files {
			processFile(file.Name(), file)
		}
	}

	return eventTxts, culledEventTxts
}
