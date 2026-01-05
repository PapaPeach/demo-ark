package demoio

import (
	"demo-ark/demoark/internal/util"
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

// Game type number values
const tournament = 0
const casual = 1
const community = 2
const mvm = 3
const valvecomp = 4

// Game type keys
const keyCulled = "culled"
const keyTournament = "tournament"
const keyCasual = "casual"
const keyCommunity = "community"
const keyMvm = "mvm"
const keyValveComp = "valvecomp"

// Name of culled directory
const culledDir = "demos_culled"

/* Zips a folder of demos from. */
func ZipDir(dirName string) {
	fmt.Printf("Zipping %s...\n", dirName)

	// Check if dirName.zip already exists
	hasConflict := false
	oldZip := dirName + "_old.zip"
	_, err := os.Stat(dirName + ".zip")
	if err == nil {
		fmt.Printf("Detected existing %s.zip, consolidating contents...\n", dirName)
		er := os.Rename(dirName+".zip", oldZip)
		if er != nil {
			e := errors.New("Error renaming existing zip: " + er.Error())
			util.EnterToExit(false, e)
		}
		hasConflict = true
	}

	// Create zip file
	zipFile, err := os.Create(dirName + ".zip")
	if err != nil {
		er := errors.New("Error creating zip file: " + err.Error())
		util.EnterToExit(false, er)
	}
	defer zipFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Copy contents of pre-existing zip to current zip
	if hasConflict {
		// Open existing zip archive for reading
		var zipReader *zip.ReadCloser
		zipReader, err = zip.OpenReader(oldZip)
		if err != nil {
			er := errors.New("Error opening existing zip: " + err.Error())
			util.EnterToExit(false, er)
		}

		// Copy contents from existing archive to new archive
		for _, file := range zipReader.File {
			err = zipWriter.Copy(file)
			if err != nil {
				er := errors.New("Error copying existing zip: " + err.Error())
				util.EnterToExit(false, er)
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
		er := errors.New("Error zipping directory: " + err.Error())
		util.EnterToExit(false, er)
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
		er := errors.New("Error opening current directory for date gathering: " + err.Error())
		util.EnterToExit(false, er)
	}
	defer directory.Close()

	contents, err := directory.Readdirnames(0)
	if err != nil {
		er := errors.New("Error reading contents of current directory for date gathering: " + err.Error())
		util.EnterToExit(false, er)
	}

	// If ZipOlderThan is 1 year, wait until Spring season (Feb) to zip last year
	year := time.Now().Year()
	if zipOlderThan == 1 && time.Now().Month() < 2 {
		year--
	}

	// Check if file is older than threshhold
	var hasSpoken bool
	for _, filename := range contents {
		// DataMajorDir=true file structure (demos_2YYY/gametype/blah.dem)
		if len(filename) == 10 && strings.HasPrefix(filename, "demos_2") {
			// Parse file's year
			ignored := ""
			fileYear := 0
			_, err = fmt.Sscanf(filename, "%6s%d", &ignored, &fileYear)
			if err != nil {
				log.Println("Error parsing year from filename:", err)
				continue
			}

			// Check if we should zip
			if year-fileYear >= int(zipOlderThan) {
				if !hasSpoken {
					fmt.Println("\nZipping takes ~0.5 seconds per demo. Please be patient")
					hasSpoken = true
				}
				ZipDir(filename)
			}
		} else if filename == "demos_tournament" ||
			filename == "demos_casual" ||
			filename == "demos_community" ||
			filename == "demos_mvm" ||
			filename == "demos_valvecomp" {
			// DateMajorDir=false file structure (demos_gametype/YYYY/blah.dem)
			// Open gametype directory
			var gameTypeDir *os.File
			gameTypeDir, err = os.Open(filename)
			if err != nil {
				log.Println("Error opening gametype directory:", err)
				continue
			}

			var gameTypeContents []string
			gameTypeContents, err = gameTypeDir.Readdirnames(0)
			if err != nil {
				log.Println("Error reading contents of gametype directory:", err)
				continue
			}
			gameTypeDir.Close()

			// Check if we should zip inner files
			for _, innerFilename := range gameTypeContents {
				// Skip impossible years
				if len(innerFilename) != len("2025") {
					continue
				}

				// Parse years from file names
				var fileYear int
				fileYear, err = strconv.Atoi(innerFilename)
				if err != nil {
					log.Println("Error parsing year from inner filename:", err)
					continue
				}

				// Check if we should zip
				if year-fileYear >= int(zipOlderThan) {
					if !hasSpoken {
						fmt.Println("\nZipping takes ~0.5 seconds per demo. Please be patient")
						hasSpoken = true
					}
					ZipDir(filename + string(filepath.Separator) + innerFilename)
				}
			}
		}
	}
}

/* Culls demos shorter than a specified minimum length. */
func CullShortDemos(demos *[]Demo, minimum uint16) []Demo {
	var cull []Demo
	keep := (*demos)[:0]
	for _, demo := range *demos {
		if demo.Duration < float32(minimum) && demo.Duration > 0.0 { // If short than minimum, move it to cull list
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

/* Culls demos of a specified game type based on a key string. */
func CullGameTypes(demos *[]Demo, key string) []Demo {
	// Parse key
	cullCasual := false
	cullCommunity := false
	cullMvm := false
	cullValveComp := false
	parsed := 0
	if strings.ContainsRune(key, 'c') {
		cullCasual = true
		parsed++
	}
	if strings.ContainsRune(key, 'q') {
		cullCommunity = true
		parsed++
	}
	if strings.ContainsRune(key, 'm') {
		cullMvm = true
		parsed++
	}
	if strings.ContainsRune(key, 'v') {
		cullValveComp = true
		parsed++
	}

	// Ensure key only contains usable characters
	if parsed != len(key) {
		err := errors.New("Invalid CullGameType key. Usage: CullGameType=cm (c = Casual, q = QuickPlay / Community, m = MvM, v = Valve Competitive).")
		util.EnterToExit(false, err)
	}

	// Mark demos of gametype for culling
	var cull []Demo
	keep := (*demos)[:0]
	for _, demo := range *demos {
		switch {
		case cullCasual && demo.GameType == casual: // Cull Casual
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: Casual\n", demo.Name)
		case cullCommunity && demo.GameType == community: // Cull Community
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\t Game Type: Community\n", demo.Name)
		case cullMvm && demo.GameType == mvm: // Cull MvM
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: MvM\n", demo.Name)
		case cullValveComp && demo.GameType == valvecomp: // Cull Valve Comp
			cull = append(cull, demo)
			fmt.Printf("Marked demo for culling: %s\tGame Type: Valve Competitive\n", demo.Name)
		default:
			keep = append(keep, demo)
		}
	}

	// Update demos list
	*demos = keep
	return cull
}

/* Generates a time format string based on arguments. */
func GetTimeFormat(twelveHourTime bool) string {
	// Format time
	if twelveHourTime {
		return "03-04-05"
	}
	return "15-04-05"
}

/* Gets the date from a demo's name. */
func GetDateTime(demo Demo) time.Time {
	// Remove leading file path
	var name string
	pathIndex := strings.LastIndex(demo.Name, string(filepath.Separator))
	if pathIndex != -1 {
		name = demo.Name[pathIndex+1:]
	} else {
		name = demo.Name
	}

	// Search for valid date my locating year via: prefix[20]YY-MM-DD_HH-MM-SS
	skipped := 0
	for i := strings.Index(name, "20"); i > -1; i = strings.Index(name[skipped:], "20") {
		dateIndex := skipped + i
		skipped += i + 1

		titleDateTime, err := time.Parse("2006-01-02_15-04-05", name[dateIndex:dateIndex+19])
		if err != nil { // This might error. We keep searching or fallback to demo's ModTime
			fmt.Println(err)
			continue
		}

		return titleDateTime
	}

	// Fallback value
	return demo.DateTime
}

/* Generates new names for demos to according to the arguments provided. */
func GetNewName(demo *Demo, dateTimeFormat string, keepPrefix bool, renameMap bool, renameDuration bool) {
	// Remove leading file path
	var name string
	pathIndex := strings.LastIndex(demo.Name, string(filepath.Separator))
	if pathIndex != -1 {
		name = demo.Name[pathIndex+1:]
	} else {
		name = demo.Name
	}

	// Get prefix
	var wishName string
	if keepPrefix {
		// Locate prefix by indexing off year
		yearString := strconv.Itoa(demo.DateTime.Year())
		if yearIdx := strings.Index(name, yearString); yearIdx != -1 {
			wishName = name[:yearIdx] + "_"
		} else {
			demIdx := strings.Index(name, ".dem")
			wishName = name[:demIdx] + "_"
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

/* Get game type of demo. */
func GetGameType(demo *Demo, showConVars bool) {
	// Determine if MvM via map prefix in header
	if strings.HasPrefix(demo.Map, "mvm_") {
		demo.GameType = mvm
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
		// Check if Casual or Valve Comp
		if CheckConVar(sc.ConVars, "mp_tournament", "1") &&
			CheckConVar(sc.ConVars, "mp_tournament_readymode", "1") &&
			CheckConVar(sc.ConVars, "mp_tournament_readymode_min", "0") {
			if CheckConVar(sc.ConVars, "mp_tournament_stopwatch", "0") &&
				CheckConVar(sc.ConVars, "sv_vote_issue_kick_allowed", "1") { // Casual
				demo.GameType = casual
				return
			} else if CheckConVar(sc.ConVars, "mp_forceautoteam", "1") { // Valve comp
				demo.GameType = valvecomp
				return
			}
		} else if !CheckConVar(sc.ConVars, "mp_tournament", "1") { // Community
			demo.GameType = community
			return
		}
	}

	// Assume file is tournament type if it's not the others
	demo.GameType = tournament
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
		demo := Demo{Name: path, NewName: file.Name(), Map: header.MapName, DateTime: fileInfo.ModTime(), Duration: header.PlaybackTime, GameType: -1}
		demos = append(demos, demo)
	}

	// Search subdirectories
	if searchDirs {
		err := filepath.WalkDir(".", func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error reading %v: %v\n", path, err)
				return nil
			}

			// Skip specified directories
			if file.IsDir() {
				// Skip ignored words
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
					if suffix == keyCulled ||
						suffix == keyTournament ||
						suffix == keyCasual ||
						suffix == keyCommunity ||
						suffix == keyMvm ||
						suffix == keyValveComp {
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

	return demos
}

func SnipeDemo(filename string) []Demo {
	// Open demo and parse info
	file, err := os.Open(filename)
	if err != nil {
		erString := fmt.Sprintf("Error opening %v: %v\n", filename, err)
		er := errors.New(erString)
		util.EnterToExit(false, er)
	}
	defer file.Close()

	fileInfo, err := os.Stat(filename)
	if err != nil {
		erString := fmt.Sprintf("Error opening %v: %v\n", filename, err)
		er := errors.New(erString)
		util.EnterToExit(false, er)
	}

	header := parser.ReadHeader(file)
	demo := Demo{Name: file.Name(), NewName: file.Name(), Map: header.MapName, DateTime: fileInfo.ModTime(), Duration: header.PlaybackTime, GameType: -1}
	demos := []Demo{demo}

	return demos
}

/* Checks if conVar exists with the desired value in a array conVars. */
func CheckConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/* Moves files in a list of files to a directory. */
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
				case tournament:
					gameType = keyTournament
				case casual:
					gameType = keyCasual
				case community:
					gameType = keyCommunity
				case mvm:
					gameType = keyMvm
				case valvecomp:
					gameType = keyValveComp
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
				util.EnterToExit(false, err)
			}

			// Move demo to directory and rename
			err = os.Rename(demos[i].Name, filepath.Join(demos[i].WishDir, demos[i].NewName))
			if err != nil {
				log.Println("Error moving demo to directory and renaming:", err)
			}
		}
	}

cull:
	// Remove previously culled for two-stage cull
	if cullMode == 1 {
		err := os.RemoveAll(filepath.Join(".", culledDir))
		if err != nil {
			util.EnterToExit(false, err)
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
			util.EnterToExit(false, err)
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
			log.Println("Error moving demo to culled directory:", err)
		}
	}
}
