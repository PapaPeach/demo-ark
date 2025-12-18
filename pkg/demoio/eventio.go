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

	// Search subdirectories
	if searchDirs {
		filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error reading %v: %v\n", path, err)
				return nil
			}

			if file.IsDir() {
				// Skip directories containing ignored words
				for _, ignoreWord := range ignoreWords {
					if strings.Contains(strings.ToLower(file.Name()), ignoreWord) {
						return fs.SkipDir
					}
				}

				// Skip already sorted directories
				// demos_YYYY
				var year int
				sorted, _ := fmt.Sscanf(file.Name(), "demos_%d", &year)
				if sorted != 0 {
					return fs.SkipDir
				}

				// Skip known demos_[known sorted]
				if len(file.Name()) >= len("demos_mvm") {
					suffix := file.Name()[6:] // demos_[suffix]
					if suffix == "culled" || suffix == "tournament" || suffix == "casual" || suffix == "mvm" {
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

	// Search subdirectories
	if searchDirs {
		filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error reading %v: %v\n", path, err)
				return nil
			}

			if file.IsDir() {
				// Skip directories containing ignored words
				for _, ignoreWord := range ignoreWords {
					if strings.Contains(strings.ToLower(file.Name()), ignoreWord) {
						return fs.SkipDir
					}
				}

				// Skip already sorted directories
				// demos_YYYY
				var year int
				sorted, _ := fmt.Sscanf(file.Name(), "demos_%d", &year)
				if sorted != 0 {
					return fs.SkipDir
				}

				// Skip known demos_[known sorted]
				if len(file.Name()) >= len("demos_mvm") {
					suffix := file.Name()[6:] // demos_[suffix]
					if suffix == "culled" || suffix == "tournament" || suffix == "casual" || suffix == "mvm" {
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

func UpdateEventTxts(eventTxts map[string][]string, culledEventTxts []string, demos []Demo, cullMode uint8) {
	// Check if a demo is in an _events.txt file
	for eventTxt, eventDemos := range eventTxts {
		// Read file
		eventFile, err := os.ReadFile(eventTxt)
		if err != nil {
			log.Println("Error reading event file for updating:", err)
			os.Exit(1)
		}
		contents := string(eventFile)

		// Look for events corresponding to demos
		for _, demo := range demos {
			demoTitle := strings.TrimSuffix(demo.Name, ".dem") // Name without .dem
			// If there is no event matching a demo, skip
			matchIndex := slices.Index(eventDemos, demoTitle)
			if matchIndex == -1 {
				continue
			}

			// Update _events.txt
			newDemoTitle := strings.TrimSuffix(demo.NewName, ".dem")
			contents = strings.ReplaceAll(contents, eventDemos[matchIndex], filepath.Join(demo.WishDir, newDemoTitle))
		}

		// Write updated contents to temporary file
		tempFile, err := os.CreateTemp(filepath.Dir(eventTxt), "temp*")
		if err != nil {
			log.Println("Error creating temporary _events.txt file:", err)
			os.Exit(1)
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
		if err != nil {
			log.Println("Error renaming temporary file:", err)
			os.Exit(1)
		}

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
		// Delete culled events
		if cullMode == 2 {
			err := os.Remove(eventTxt)
			if err != nil {
				log.Println("Error deleting culled _events.txt:", err)
			}
			continue
		}

		// Move event to culled directory
		err := os.Rename(eventTxt, filepath.Join(culledDir, eventTxt))
		if err != nil {
			log.Println(err)
		}
	}
}

func UpdateEventJsons(eventJsons []string, culledEventJsons []string, demos []Demo, cullMode uint8) {
	// Search for json corresponding to demos
	for _, eventJson := range eventJsons {
		foundMatch := false
		for _, demo := range demos {
			targetTitle := strings.Replace(demo.Name, ".dem", ".json", 1)
			if eventJson != targetTitle {
				continue
			}

			// Update matching json name
			foundMatch = true
			newName := strings.Replace(demo.NewName, ".dem", ".json", 1)
			err := os.Rename(eventJson, filepath.Join(demo.WishDir, newName))
			if err != nil {
				log.Println("Error updating .json:", err)
				os.Exit(1)
			}
		}

		// Cull jsons with no corresponding demo
		if !foundMatch {
			culledEventJsons = append(culledEventJsons, eventJson)
		}
	}

	// If there's no events to cull, skip culling
	culledDir := "demos_culled"
	length := len(culledEventJsons)
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
	for _, eventJson := range culledEventJsons {
		// Delete culled events
		if cullMode == 2 {
			err := os.Remove(eventJson)
			if err != nil {
				log.Println("Error deleting culled .json:", err)
			}
			continue
		}

		// Move event to culled directory
		err := os.Rename(eventJson, filepath.Join(culledDir, eventJson))
		if err != nil {
			log.Println(err)
		}
	}
}
