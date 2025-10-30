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
	"strconv"
	"strings"
	"sync"
)

type Demo struct {
	Name     string
	Year     int
	Duration float32
	Map      string
}

type DemoYear struct {
	Year       int
	Casual     []Demo
	Mvm        []Demo
	Tournament []Demo
}

type Arguments struct {
	Silent         bool     // Run program without prompts
	SortYear       bool     // Group demos by year
	SortGameType   bool     // Group demos by game type
	RenameMap      bool     // Rename the demo to contain the map name
	RenameDate     bool     // Rename the demo to contain the recorded/edited date
	RenameTime     bool     // Rename the demo to contain the recored/edited time
	RenameDuration bool     // Rename demo to contain the duration of the demo
	SearchFolders  bool     // Search for folders within the current directory
	Multithread    bool     // Allow the use of multiple cores / threads
	ZipOlderThan   uint8    // Zip demos older than this many years
	CullBelow      uint8    // Number of seconds that demos below that duration will be deleted
	MajorDirectory uint8    // Ex: year/gametype/demo.dem | gametype/year/demo.dem
	IgnoreWords    []string // Ignore file / folder names containing string
}

/** Parses boolean arguments and returns the boolean value */
func parseBoolArg(args []string, str string, message string) bool {
	if i := slices.Index(args, str); i != -1 {
		switch args[i+1] {
		case "1":
			fallthrough
		case "true":
			fmt.Println(message)
			return true
		case "0":
			fallthrough
		case "false":
			return false
		default:
			log.Printf("Invalid argument value: %s = %s\n", args[i], args[i+1])
			enterToExit(false)
		}
	}

	return false
}

/** Parses integer arguments and returns the integer value */
func parseIntArg(args []string, str string, message string) uint8 {
	if i := slices.Index(args, str); i != -1 {
		value, err := strconv.Atoi(args[i+1])
		if err != nil {
			log.Printf("Invalid argument value: %s = %s\n", args[i], args[i+1])
			enterToExit(false)
		}

		// If value wouldn't fit
		if value > 255 {
			log.Printf("Invalid argument value: %s = %s\tMaximum value: 255", args[i], args[i+1])
			enterToExit(false)
		}

		fmt.Println(message)
		return uint8(value)
	}

	return 0
}

/** Gets argument values */
func getArgs() Arguments {
	const Silent = "silent"
	const SortYear = "sortyear"
	const SortGameType = "sortgametype"
	const RenameMap = "renamemap"
	const RenameDate = "renamedate"
	const RenameTime = "renametime"
	const RenameDuration = "renameduration"
	const SearchFolders = "searchfolders"
	const Multithread = "multithread"
	const ZipOlderThan = "zipolderthan"
	const CullBelow = "cullbelow"
	const MajorDirectory = "majordirectory"
	const IgnoreWords = "ignorewords"

	// Convert args to lower case
	var args []string
	for _, arg := range os.Args[1:] {
		args = append(args, strings.ToLower(arg))
	}

	// Get bool arg values
	var a Arguments
	a.Silent = parseBoolArg(args, Silent, "Running silently")
	a.SortYear = parseBoolArg(args, SortYear, "Sorting years")
	a.SortGameType = parseBoolArg(args, SortGameType, "Sorting game types")
	a.RenameMap = parseBoolArg(args, RenameMap, "Renaming with map name")
	a.RenameDate = parseBoolArg(args, RenameDate, "Renaming with record date")
	a.RenameTime = parseBoolArg(args, RenameTime, "Renaming with record time")
	a.RenameDuration = parseBoolArg(args, RenameDuration, "Rename with demo duration")
	a.SearchFolders = parseBoolArg(args, SearchFolders, "Searching folders")
	a.Multithread = parseBoolArg(args, Multithread, "Multithreading enabled")

	// Get int arg values
	a.ZipOlderThan = uint8(parseIntArg(args, ZipOlderThan, "Zipping old demos"))
	a.CullBelow = uint8(parseIntArg(args, CullBelow, "Culling short demos"))

	// Get major directory
	if i := slices.Index(args, MajorDirectory); i != -1 {
		switch args[i+1] {
		case "year":
			a.MajorDirectory = 0
		case "gametype":
			a.MajorDirectory = 1
		default:
			log.Printf("Invalid argument value: %s = %s\n", args[i], args[i+1])
			enterToExit(false)
		}
	}

	// Get ignore words
	if i := slices.Index(args, IgnoreWords); i != -1 {
		a.IgnoreWords = append(a.IgnoreWords, args[i+1:]...)
	}

	return a
}

/** Checks if conVar exists with the desired value in a array conVars */
func checkConVar(conVars []string, wishStr string, wishVal string) bool {
	if i := slices.Index(conVars, wishStr); i != -1 && conVars[i+1] == wishVal {
		return true
	}
	return false
}

/** Returns a list of .dem files in the current directory */
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

/** Group demos by gametype */
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

/** Groups demos by year */
func groupYears(demos []Demo) map[int][]Demo {
	years := make(map[int][]Demo)
	for _, demo := range demos {
		years[demo.Year] = append(years[demo.Year], demo)
	}

	return years
}

/** Moves files in a list of files to a directory */
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

/** Prompts user to press enter to exit the program, then exit when entered */
func enterToExit(silent bool) {
	if !silent {
		fmt.Println("Press enter to close.")
		fmt.Scanln()
	}
	os.Exit(0)
}

func main() {
	args := getArgs()

	// Get demos
	demoList := getDemos()
	demoListCount := len(demoList)
	fmt.Printf("Scanning %d demos...\n", demoListCount)

	//fmt.Println(time.Now().Year())
	years := groupYears(demoList)

	for year, demos := range years {
		// Asynchronously scan demos if there's a lot
		var casualDemos []Demo
		var mvmDemos []Demo
		var tournamentDemos []Demo
		demoCount := len(demos)
		cores := (runtime.NumCPU() / 2) - 1 // Avoid e-cores and keep one core open
		if cores > 1 && demoCount > cores*10 {
			type GameTypes struct {
				Casual     []Demo
				Mvm        []Demo
				Tournament []Demo
			}

			// Determine how to split demos list to distribute across cores
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
				casualDemos = append(casualDemos, scanned[i].Casual...)
				mvmDemos = append(mvmDemos, scanned[i].Mvm...)
				tournamentDemos = append(tournamentDemos, scanned[i].Tournament...)
			}
		} else { // Scan using single core
			casualDemos, mvmDemos, tournamentDemos = groupGameTypes(demos)
		}

		// Move files to corresponding directories
		dirPrefix := filepath.Join(strconv.Itoa(year)+"demos", "")
		var movers sync.WaitGroup
		movers.Go(func() { moveToDirectory(dirPrefix+"casual", casualDemos) })
		movers.Go(func() { moveToDirectory(dirPrefix+"mvm", mvmDemos) })
		movers.Go(func() { moveToDirectory(dirPrefix+"tournament", tournamentDemos) })
		movers.Wait()
	}

	enterToExit(args.Silent)
}
