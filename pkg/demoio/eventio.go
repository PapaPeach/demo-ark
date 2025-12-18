package demoio

import (
	"bufio"
	"errors"
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

		// If file is empty, cull it
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
					demoName = scanner.Text()[demoIndex : demoIndex+endIndex]
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

/* Get event jsons associated with demos */
func GetEventJsons(searchDirs bool, ignoreWords []string) ([]string, []string) {
	// Filter list to only have .json files
	var eventJsons []string
	var culledEventJsons []string
	processFile := func(path string, file fs.DirEntry) {
		// Skip irrelevant files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			return
		}

		// If file is empty, cull it
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Printf("Error getting stats on %v: %v\n", path, err)
		}
		if fileInfo.Size() == 0 {
			culledEventJsons = append(culledEventJsons, path)
			return
		}

		eventJsons = append(eventJsons, path)
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

	return eventJsons, culledEventJsons
}

func UpdateEventTxts(eventTxts map[string][]string, demos []Demo, culledEventTxts []string, cullMode uint8) {
	// Check if a demo is in an _events.txt file
	for eventTxt, eventDemos := range eventTxts {
		// Read file
		eventFile, err := os.ReadFile(eventTxt)
		if err != nil {
			log.Println("Error reading event file for updating:", err)
		}
		contents := string(eventFile)

		// Look for events corresponding to demos
		for _, demo := range demos {
			demoTitle := strings.TrimSuffix(demo.Name, ".dem") // Name without .dem
			fmt.Println("Searching for: ", demoTitle)
			// If there is no event matching a demo, skip
			targetIndex := slices.Index(eventDemos, demoTitle)
			if targetIndex == -1 {
				continue
			}

			// Update _events.txt
			newDemoTitle := strings.TrimSuffix(demo.NewName, ".dem")
			fmt.Printf("Replacing: %s\tWith: %s\n", eventDemos[targetIndex], filepath.Join(demo.WishDir, newDemoTitle))
			contents = strings.ReplaceAll(contents, eventDemos[targetIndex], filepath.Join(demo.WishDir, newDemoTitle))
		}

		// Write updated contents to temporary file
		tempFile, err := os.CreateTemp(filepath.Dir(eventTxt), "temp*")
		if err != nil {
			log.Println("Error creating temporary _events.txt file:", err)
		}

		_, err = tempFile.WriteString(contents)
		if err != nil {
			log.Println("Error writing to temporary _events.txt file:", err)
			tempFile.Close()
			os.Exit(1)
		}
		tempFile.Close()

		// Rename temp file to _events.txt
		err = os.Rename(tempFile.Name(), eventTxt)

		// Remove temp file?
		os.Remove(tempFile.Name())
	}

	// If there's no events to cull, skip culling
	culledDir := "demos_culled"
	length := len(culledEventTxts)
	if length == 0 {
		return
	}

	// Create the culled directory
	if cullMode < 2 {
		// Make directory to move demo to
		err := os.MkdirAll(culledDir, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			log.Println(err)
			os.Exit(1)
		}
	}

	// Cull demos
	for _, eventTxt := range culledEventTxts {
		// Delete culled demos
		if cullMode == 2 {
			err := os.Remove(eventTxt)
			if err != nil {
				log.Println("Error deleting culled _events.txt:", err)
			}
			continue
		}

		// Move demo to culled directory
		err := os.Rename(eventTxt, filepath.Join(culledDir, eventTxt))
		if err != nil {
			log.Println(err)
		}
	}
}
