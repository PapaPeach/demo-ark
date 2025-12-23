package demoio

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/* Get screenshots from ds_screens associated with demos. */
func GetScreenshots(searchDirs bool, ignoreWords []string) []string {
	// Filter list to only have screenshots
	var screenshots []string
	processFile := func(path string, file fs.DirEntry) {
		// Skip irrelevant files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tga") {
			return
		}

		screenshots = append(screenshots, path)
	}

	// Search subdirectories
	if searchDirs {
		// Prepend the screenshots/.. directory for skipping
		ignoreWords = append([]string{"screenshots"}, ignoreWords...)
		eventWalker(ignoreWords, processFile)
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

	return screenshots
}

/* Update screenshots with their demos or cull demoless screenshots. */
func UpdateScreenshots(screenshots []string, demos []Demo, cullScreenshots bool, cullMode uint8) {
	// Search for screenshots corresponding to demos
	var culledScreenshots []string
	for _, screenshot := range screenshots {
		// If we're culling all, skip processing
		if cullScreenshots {
			culledScreenshots = append(culledScreenshots, screenshot)
			continue
		}

		// Look for demos matching screenshot
		foundMatch := false
		for _, demo := range demos {
			targetTitle := strings.Replace(demo.Name, ".dem", ".tga", 1)
			if screenshot != targetTitle {
				continue
			}

			// Update matching screenshot name
			foundMatch = true
			newName := strings.Replace(demo.NewName, ".dem", ".tga", 1)
			err := os.Rename(screenshot, filepath.Join(demo.WishDir, newName))
			if err != nil {
				log.Println("Error updating .tga:", err)
				os.Exit(1)
			}
		}

		// Cull screenshots with no corresponding demo
		if !foundMatch {
			fmt.Printf("Culled screenshot without corresponding demo: %s\n", screenshot)
			culledScreenshots = append(culledScreenshots, screenshot)
		}
	}

	// If there's no screenshots to cull, skip culling
	length := len(culledScreenshots)
	if length == 0 {
		return
	}

	// Handle length accordingly
	if length == 1 {
		fmt.Println("Culling 1 screenshot...")
	} else { // Plural demos
		fmt.Printf("Culling %d screenshots...\n", length)
	}

	// Create the culled directory
	if cullMode < 2 {
		// Make directory to move screenshots to
		err := os.MkdirAll(culledDir, os.ModePerm)
		if err != nil && !errors.Is(err, os.ErrExist) {
			log.Println(err)
			os.Exit(1)
		}
	}

	// Cull screenshots
	for _, screenshot := range culledScreenshots {
		// Delete culled screenshots
		if cullMode == 2 {
			err := os.Remove(screenshot)
			if err != nil {
				log.Println("Error deleting culled screenshot:", err)
			}
			continue
		}

		// Move event to culled directory
		newName := strings.ReplaceAll(screenshot, string(filepath.Separator), ".")
		err := os.Rename(screenshot, filepath.Join(culledDir, newName))
		if err != nil {
			log.Println(err)
		}
	}
}
