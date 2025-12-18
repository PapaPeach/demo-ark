package demoio

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	zip "github.com/klauspost/compress/zip"
	parser "github.com/papapeach/tf-demo-parser"
)

type Demo struct {
	Name     string    // Current file name of demo
	NewName  string    // Desired rename for demo
	WishDir  string    // Desired directory for demo after sorting
	Map      string    // Map demo was recorded on
	DateTime time.Time // Date and time demo was recorded / edited
	Duration float32   // Duration of demo in seconds
	GameType int8      // 0: Tournament | 1: Casual | 2: MvM
}

/* Zips a folder of demos from */
func ZipDir(dirName string) {
	fmt.Printf("Zipping %s...\n", dirName)

	// Check if dirName.zip already exists
	hasConflict := false
	oldZip := dirName + "_old.zip"
	_, err := os.Stat(dirName + ".zip")
	if err == nil {
		fmt.Printf("Detected existing %s.zip, consolidating contents...\n", dirName)
		os.Rename(dirName+".zip", oldZip)
		hasConflict = true
	}

	// Create zip file
	zipFile, err := os.Create(dirName + ".zip")
	if err != nil {
		log.Println("Error creating zip file:", err)
		os.Exit(1)
	}
	defer zipFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Copy contents of pre-existing zip to current zip
	if hasConflict {
		// Open existing zip archive for reading
		zipReader, err := zip.OpenReader(oldZip)
		if err != nil {
			log.Println("Error opening existing zip:", err)
			os.Exit(1)
		}

		// Copy contents from existing archive to new archive
		for _, file := range zipReader.File {
			err = zipWriter.Copy(file)
			if err != nil {
				log.Println("Error copying existing zip:", err)
				os.Exit(1)
			}
		}
		zipReader.Close()

		// Remove old zip
		err = os.Remove(oldZip)
		if err != nil {
			log.Println("Error removing existing zip:", err)
		}
	}

	// Zip contents of directory
	dir := os.DirFS(dirName)
	err = zipWriter.AddFS(dir)
	if err != nil {
		log.Println("Error zipping directory:", err)
		os.Exit(1)
	}

	// Removed source directory
	err = os.RemoveAll(dirName)
	if err != nil {
		log.Println("Error removing directory after zipping:", err)
	}
}

func ZipOldDemos(zipOlderThan uint8) {
	// Read contents of current directory for "demos_..."
	directory, err := os.Open(".")
	if err != nil {
		log.Println("Error opening current directory for date gathering:", err)
		os.Exit(1)
	}
	defer directory.Close()

	contents, err := directory.Readdirnames(0)
	if err != nil {
		log.Println("Error reading contents of current directory for date gathering:", err)
		os.Exit(1)
	}

	// Check if file is older than threshhold
	year := time.Now().Year()
	for _, filename := range contents {
		// DataMajorDir=true file structure (demos_2YYY/gametype/blah.dem)
		if len(filename) == 10 && strings.HasPrefix(filename, "demos_2") {
			// Parse file's year
			ignored := ""
			fileYear := 0
			_, err := fmt.Sscanf(filename, "%6s%d", &ignored, &fileYear)
			if err != nil {
				log.Println("Error parsing year from filename:", err)
				continue
			}

			// Check if we should zip
			if year-fileYear >= int(zipOlderThan) {
				ZipDir(filename)
			}
		} else if filename == "demos_tournament" || filename == "demos_casual" || filename == "demos_mvm" {
			// DateMajorDir=false file structure (demos_gametype/YYYY/blah.dem)
			// Open gametype directory
			gameTypeDir, err := os.Open(filename)
			if err != nil {
				log.Println("Error opening gametype directory:", err)
				continue
			}

			gameTypeContents, err := gameTypeDir.Readdirnames(0)
			if err != nil {
				log.Println("Error reading contents of gametype directory:", err)
				continue
			}
			gameTypeDir.Close()

			// Check if we should zip inner files
			for _, innerFilename := range gameTypeContents {
				// Skip impossible years
				if len(innerFilename) != 4 {
					continue
				}

				// Parse years from file names
				fileYear, err := strconv.Atoi(innerFilename)
				if err != nil {
					log.Println("Error parsing year from inner filename:", err)
					continue
				}

				// Check if we should zip
				if year-fileYear >= int(zipOlderThan) {
					ZipDir(filename + string(filepath.Separator) + innerFilename)
				}
			}
		}
	}
}

/* Culls demos shorter than a specified minimum length */
func CullShortDemos(demos *[]Demo, min uint16) []Demo {
	var cull []Demo
	keep := (*demos)[:0]
	for _, demo := range *demos {
		if demo.Duration < float32(min) && demo.Duration > 0.0 { // If short than minimum, move it to cull list
			cull = append(cull, demo)
			fmt.Printf("Marked short demo for culling: %s\tDuration: %.2f seconds\n", demo.Name, demo.Duration)
		} else { // If longer than minimum, keep it in main demo list
			keep = append(keep, demo)
			if demo.Duration == 0.0 {
				fmt.Printf("%s reported a duration of 0.00 seconds.\nThis indicates a TF2 bug occured while recording. Demo will will not be culled.\n", demo.Name)
			}
		}
	}

	// Update demos list
	*demos = keep
	return cull
}

/* Culls demos of a specified game type based on a key string */
func CullGameTypes(demos *[]Demo, key string) []Demo {
	// Parse key
	cullTournament := false
	cullCasual := false
	cullMvm := false
	parsed := 0
	if strings.ContainsRune(key, 't') {
		cullTournament = true
		parsed++
	}
	if strings.ContainsRune(key, 'c') {
		cullCasual = true
		parsed++
	}
	if strings.ContainsRune(key, 'm') {
		cullMvm = true
		parsed++
	}

	// Ensure key only contains usable characters
	if parsed != len(key) {
		log.Printf("Invalid CullGameType key. Usage: CullGameType=tcm (t = Tournament, c = Casual, m = MvM).\n")
		os.Exit(1)
	}

	// Mark demos of gametype for culling
	var cull []Demo
	keep := (*demos)[:0]
	for _, demo := range *demos {
		if cullTournament && demo.GameType == 0 { // Cull Tournament
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: Tournament\n", demo.Name)
		} else if cullCasual && demo.GameType == 1 { // Cull Casual
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: Casual\n", demo.Name)
		} else if cullMvm && demo.GameType == 2 { // Cull MvM
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: MvM\n", demo.Name)
		} else { // Don't mark for culling
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
func GetNewName(demo *Demo, dateTimeFormat string, keepPrefix bool, renameMap bool, renameDuration bool) {
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
		wishName = fmt.Sprintf("%s_%02d-%02d", wishName, length/60, length%60)
	}

	// Add .dem and set demo's new name
	demo.NewName = wishName + ".dem"
}

/* Get game type of demo (0: Tournament | 1: Casual | 2: MvM) */
func GetGameType(demo *Demo, showConVars bool) {
	// Determine if MvM via map prefix in header
	if strings.HasPrefix(demo.Map, "mvm_") {
		demo.GameType = 2
		return
	}

	// Get message contents
	file, err := os.Open(demo.Name)
	if err != nil {
		fmt.Printf("Error opening %v: %v", demo.Name, err)
		return
	}
	defer file.Close()
	file.Seek(1072, io.SeekStart) // Skip header of known length
	msg := parser.ReadMessage(file)

	// Prints parsed ConVars for ShowConVars=true
	if showConVars {
		fmt.Println("Showing console variables for:", demo.Name)
		for _, sc := range msg.ParsedData.SetConVar {
			parser.PrintSetConVar(sc)
		}
		fmt.Println()
	}

	// Determine if casual via specific conVar values
	for _, sc := range msg.ParsedData.SetConVar {
		// Check if tournament 1, stopwatch 0, readymode 1, readymode_min 0
		if CheckConVar(sc.ConVars, "mp_tournament", "1") &&
			CheckConVar(sc.ConVars, "mp_tournament_stopwatch", "0") &&
			CheckConVar(sc.ConVars, "mp_tournament_readymode", "1") &&
			CheckConVar(sc.ConVars, "mp_tournament_readymode_min", "0") {
			demo.GameType = 1
			return
		}
	}

	// Assume file is tournament type if it's not the others
	demo.GameType = 0
}

func GetDemos(searchDirs bool, ignoreWords []string) []Demo {
	// Filter list to only have .dem files
	var demos []Demo
	processFile := func(path string, file fs.DirEntry) {
		// Skip non-demo files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".dem") {
			return
		}

		// Skip ignored words
		for _, ignoreWord := range ignoreWords {
			if strings.Contains(strings.ToLower(file.Name()), ignoreWord) {
				return
			}
		}

		// Get header for map name and duration
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error opening %v: %v\n", path, err)
		}
		defer f.Close()

		header := parser.ReadHeader(f)

		fileInfo, err := f.Stat()
		if err != nil {
			fmt.Printf("Error getting info on %v: %v", path, err)
		}
		demo := Demo{Name: path, NewName: file.Name(), Map: header.MapName, DateTime: fileInfo.ModTime(), Duration: header.PlaybackTime, GameType: 0}
		demos = append(demos, demo)
	}

	if searchDirs { // Search subdirectories
		filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error reading %v: %v\n", path, err)
				return nil
			}

			// Skip already sorted directories
			if file.IsDir() {
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

	return demos
}

func SnipeDemo(filename string) []Demo {
	// Open demo and parse info
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening %v: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Error opening %v: %v\n", filename, err)
		os.Exit(1)
	}

	header := parser.ReadHeader(file)
	demo := Demo{Name: file.Name(), NewName: file.Name(), Map: header.MapName, DateTime: fileInfo.ModTime(), Duration: header.PlaybackTime, GameType: -1}
	demos := []Demo{demo}

	return demos
}

/* Checks if conVar exists with the desired value in a array conVars */
func CheckConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/* Moves files in a list of files to a directory */
func SortDemos(demoList *[]Demo, culled []Demo, sortYear bool, sortGameType bool, dateMajorDir bool, showConVars bool, cullMode uint8) {
	demos := *demoList
	// Handle length accordingly
	switch length := len(demos); length {
	case 0: // Skip to culling if no demos to sort
		goto cull
	case 1: // Singular demo file
		fmt.Printf("Sorting %d demo...\n", length)
	default: // Plural demos
		fmt.Printf("Sorting %d demos...\n", length)
	}

	// Sort demos into year and/or gametype
	if sortYear || sortGameType {
		// Check if we have already gotten game types
		hasGameTypes := false
		if demos[0].GameType != -1 {
			hasGameTypes = true
		}

		for i := range demos {
			// Get game type for SortGameType=true
			gameType := ""
			if sortGameType {
				// Only run GetGameType if we don't already have it
				if !hasGameTypes {
					GetGameType(&demos[i], showConVars)
				}
				switch demos[i].GameType {
				case 0:
					gameType = "tournament"
				case 1:
					gameType = "casual"
				case 2:
					gameType = "mvm"
				}
			}

			// Get year as a string for SortYear=true
			year := ""
			if sortYear {
				year = strconv.Itoa(demos[i].DateTime.Year())
			}

			// Get name of directory to move demo to
			demos[i].WishDir = "demos_"
			if dateMajorDir { // demos_2025/tournament/demo.dem
				demos[i].WishDir += filepath.Join(year, gameType)
			} else { // demos_tournament/2025/demo.dem
				demos[i].WishDir += filepath.Join(gameType, year)
			}

			// Make directory to move demo to
			err := os.MkdirAll(demos[i].WishDir, os.ModePerm)
			if err != nil && !errors.Is(err, os.ErrExist) {
				log.Println(err)
				os.Exit(1)
			}

			// Move demo to directory and rename
			err = os.Rename(demos[i].Name, filepath.Join(demos[i].WishDir, demos[i].NewName))
			if err != nil {
				log.Println(err)
			}
		}
	}

cull:
	// Remove previously culled for two-stage cull
	culledDir := "demos_culled"
	if cullMode == 1 {
		err := os.RemoveAll(filepath.Join(".", culledDir))
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}
	}

	// If there's no demos to cull, skip culling
	length := len(culled)
	if length == 0 {
		return
	}

	// Handle length accordingly
	if length == 1 {
		fmt.Println("Culling 1 demo...")
	} else { // Plural demos
		fmt.Printf("Culling %d demos...\n", length)
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
	for _, demo := range culled {
		// Delete culled demos
		if cullMode == 2 {
			err := os.Remove(demo.Name)
			if err != nil {
				log.Println("Error deleting culled demo:", err)
			}
			continue
		}

		// Move demo to culled directory
		err := os.Rename(demo.Name, filepath.Join(culledDir, demo.NewName))
		if err != nil {
			log.Println(err)
		}
	}
}
