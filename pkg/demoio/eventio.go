package demoio

import (
	"bufio"
	"demo-ark/demoark/internal/util"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

/* Walker function for searching directories and processing files appropriately. */
func eventWalker(ignoreWords []string, processFile func(path string, file fs.DirEntry)) {
	err := filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
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
				if suffix == "culled" ||
					suffix == "tournament" ||
					suffix == "casual" ||
					suffix == "community" ||
					suffix == "mvm" ||
					suffix == "valvecomp" {
					return fs.SkipDir
				}
			}
		}

		processFile(path, file)
		return nil
	})
	if err != nil {
		er := errors.New("Error walking directory: " + err.Error())
		util.EnterToExit(false, er)
	}
}

/* Get _event.txt files and associated demos. */
func GetEventTxts(searchDirs bool, cullEventTxts bool, ignoreWords []string) (map[string][]string, []string) {
	// Filter list to only have _events.txt files
	eventTxts := make(map[string][]string)
	var culledEventTxts []string
	processFile := func(path string, file fs.DirEntry) {
		// Skip irrelevant files
		if file.IsDir() || !strings.HasSuffix(path, "_events.txt") {
			return
		}

		// If we're culling all, skip processing
		if cullEventTxts {
			culledEventTxts = append(culledEventTxts, path)
			return
		}

		// If file is empty, cull it
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Printf("Error getting stats on %v: %v\n", path, err)
		} else if fileInfo.Size() == 0 {
			fmt.Printf("Marked empty event file for culling: %s\n", path)
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
		eventWalker(ignoreWords, processFile)
	} else { // Just search current directory
		// Get list of files in current directory
		dir, err := os.Open(".")
		if err != nil {
			util.EnterToExit(false, err)
		}
		files, err := dir.ReadDir(0)
		if err != nil {
			util.EnterToExit(false, err)
		}
		defer dir.Close()

		// Filter list to contain only demos
		for _, file := range files {
			processFile(file.Name(), file)
		}
	}

	return eventTxts, culledEventTxts
}

/* Get event jsons associated with demos. */
func GetEventJsons(searchDirs bool, cullEventJsons bool, ignoreWords []string) ([]string, []string) {
	// Filter list to only have .json files
	var eventJsons []string
	var culledEventJsons []string
	processFile := func(path string, file fs.DirEntry) {
		// Skip irrelevant files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			return
		}

		// If we're culling all, skip processing
		if cullEventJsons {
			culledEventJsons = append(culledEventJsons, path)
			return
		}

		// Check if file lacks bookmarks, cull it
		contents, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error checking contents of %v: %v", path, err)
		} else if !strings.Contains(string(contents), "\t\t") { // No bookmarks
			fmt.Printf("Marked empty event file for culling: %s\n", path)
			culledEventJsons = append(culledEventJsons, path)
			return
		}

		eventJsons = append(eventJsons, path)
	}

	// Search subdirectories
	if searchDirs {
		eventWalker(ignoreWords, processFile)
	} else { // Just search current directory
		// Get list of files in current directory
		dir, err := os.Open(".")
		if err != nil {
			util.EnterToExit(false, err)
		}
		files, err := dir.ReadDir(0)
		if err != nil {
			util.EnterToExit(false, err)
		}
		defer dir.Close()

		// Filter list to contain only demos
		for _, file := range files {
			processFile(file.Name(), file)
		}
	}

	return eventJsons, culledEventJsons
}

/* Updated _events.txt files with new demo names and cull empty _events.txt. */
func UpdateEventTxts(eventTxts map[string][]string, culledEventTxts []string, demos []Demo, cullEventTxts bool, cullMode uint8) {
	// If we're culling all, skip processing
	if cullEventTxts {
		goto culling
	}

	// Check if a demo is in an _events.txt file
	for eventTxt, eventDemos := range eventTxts {
		// Read file
		eventFile, err := os.ReadFile(eventTxt)
		if err != nil {
			er := errors.New("Error reading event file for updating: " + err.Error())
			util.EnterToExit(false, er)
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
			er := errors.New("Error creating temporary _events.txt file: " + err.Error())
			util.EnterToExit(false, er)
		}

		_, err = tempFile.WriteString(contents)
		if err != nil {
			tempFile.Close()
			er := errors.New("Error writing to temporary _events.txt file: " + err.Error())
			util.EnterToExit(false, er)
		}
		tempFile.Close()

		// Rename temp file to _events.txt
		err = os.Rename(tempFile.Name(), eventTxt)
		if err != nil {
			er := errors.New("Error renaming temporary file: " + err.Error())
			util.EnterToExit(false, er)
		}

		// TODO: Remove temp file?
		os.Remove(tempFile.Name())
	}

culling:
	// If there's no events to cull, skip culling
	length := len(culledEventTxts)
	if length == 0 {
		return
	}

	// Handle length accordingly
	if length == 1 {
		fmt.Println("Culling 1 event .json...")
	} else { // Plural demos
		fmt.Printf("Culling %d event .jsons...\n", length)
	}

	// Create the culled directory
	if cullMode < 2 {
		// Make directory to move events to
		err := os.MkdirAll(culledDir, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			util.EnterToExit(false, err)
		}
	}

	// Cull events
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
		newName := strings.ReplaceAll(eventTxt, string(filepath.Separator), ".")
		err := os.Rename(eventTxt, filepath.Join(culledDir, newName))
		if err != nil {
			log.Println(err)
		}
	}
}

/* Update .json event files and cull empty or demoless .json files. */
func UpdateEventJsons(eventJsons []string, culledEventJsons []string, demos []Demo, cullEventJsons bool, cullMode uint8) {
	// If we're culling all, skip processing
	if cullEventJsons {
		goto culling
	}

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
				er := errors.New("Error updating .json: " + err.Error())
				util.EnterToExit(false, er)
			}
		}

		// Cull jsons with no corresponding demo
		if !foundMatch {
			fmt.Printf("Culled event file without corresponding demo: %s\n", eventJson)
			culledEventJsons = append(culledEventJsons, eventJson)
		}
	}

culling:
	// If there's no events to cull, skip culling
	length := len(culledEventJsons)
	if length == 0 {
		return
	}

	// Handle length accordingly
	if length == 1 {
		fmt.Println("Culling 1 _events.txt...")
	} else { // Plural demos
		fmt.Printf("Culling %d _event.txts...\n", length)
	}

	// Create the culled directory
	if cullMode < 2 {
		// Make directory to move events to
		err := os.MkdirAll(culledDir, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			util.EnterToExit(false, err)
		}
	}

	// Cull events
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
		newName := strings.ReplaceAll(eventJson, string(filepath.Separator), ".")
		err := os.Rename(eventJson, filepath.Join(culledDir, newName))
		if err != nil {
			log.Println(err)
		}
	}
}
