package main

import (
	"demo-ark/demoark/pkg/demoio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type Demo = demoio.Demo

type Arguments struct {
	Silent         bool     // Run program without prompts
	SortYear       bool     // Group demos by year
	SortMonth      bool     // Group demos by month
	SortGameType   bool     // Group demos by game type
	KeepPrefix     bool     // Rename options won't overwrite a detected ds_prefix
	RenameMap      bool     // Rename the demo to contain the map name
	RenameDuration bool     // Rename demo to contain the duration of the demo
	SearchDirs     bool     // Search for subdirectories within the current directory
	Multithread    bool     // Allow the use of multiple cores / threads
	DateMajorDir   bool     // True: year/month/gametype/demo.dem | False: gametype/year/month/demo.dem
	UseEditDate    bool     // Use the date that a demo was last edited rather than date in its file name
	TwelveHourTime bool     // True: 12hr | false: 24hr
	SetAsideCulled bool     // Set aside culled demos to a "culled" directory, rather than deleting them
	ShowConVars    bool     // Outputs console variables parsed from demo (mainly for debugging)
	ZipOlderThan   uint8    // Zip demos older than this many years
	CullBelow      uint8    // Number of seconds that demos below that duration will be deleted
	Snipe          string   // Snipe a specific file (exactly) to execute program on (mainly for debugging)
	IgnoreWords    []string // Ignore file / folder names containing string
}

/* Parses boolean arguments and returns the boolean value */
func parseBoolArg(args []string, keyword string, message string) bool {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			switch arg[len(keyword)+1:] {
			case "1":
				fallthrough
			case "true":
				fmt.Println(strings.ToUpper(message[:1]) + message[1:])
				return true
			case "0":
				fallthrough
			case "false":
				fmt.Println("Not", message)
				return false
			default:
				log.Printf("Invalid argument value: %s\n", arg)
				enterToExit(false)
			}
		}
	}

	return false
}

/* Parses integer arguments and returns the integer value */
func parseIntArg(args []string, keyword string) uint8 {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			value, err := strconv.Atoi(arg[len(keyword)+1:])
			if err != nil {
				log.Printf("Invalid argument value: %s\n", arg)
				enterToExit(false)
			}

			// If value wouldn't fit
			if value > 255 {
				log.Printf("Invalid argument value: %s\tMaximum value: 255\n", arg)
				enterToExit(false)
			}

			return uint8(value)
		}
	}

	return 0
}

/* Gets argument values */
func getArgs() Arguments {
	const Silent = "silent"                 //
	const SortYear = "sortyear"             // TODO
	const SortMonth = "sortmonth"           // TODO
	const SortGameType = "sortgametype"     // TODO
	const KeepPrefix = "keepprefix"         //
	const RenameMap = "renamemap"           //
	const RenameDuration = "renameduration" //
	const SearchDirs = "searchdirs"         // TODO
	const Multithread = "multithread"       // TODO: Partial
	const DateMajorDir = "datemajordir"     // TODO
	const UseEditDate = "useeditdate"       //
	const TwelveHourTime = "twelvehourtime" //
	const ShowConVars = "showconvars"       // TODO
	const SetAsideCulled = "setasideculled" // TODO
	const ZipOlderThan = "zipolderthan"     // TODO
	const CullBelow = "cullbelow"           // TODO
	const Snipe = "snipe"                   // TODO
	const IgnoreWords = "ignorewords"       //

	// Convert args to lower case
	var args []string
	for _, arg := range os.Args[1:] {
		args = append(args, strings.ToLower(arg))
	}

	// Set defaults
	a := Arguments{
		Silent:         false,
		SortYear:       true,
		SortMonth:      false,
		SortGameType:   true,
		KeepPrefix:     true,
		RenameMap:      false,
		RenameDuration: false,
		SearchDirs:     false,
		Multithread:    true,
		DateMajorDir:   true,
		UseEditDate:    false,
		TwelveHourTime: false,
		ShowConVars:    false,
		SetAsideCulled: true,
		ZipOlderThan:   1,
		CullBelow:      10,
		Snipe:          "",
		IgnoreWords:    []string{"reference"},
	}

	// Get bool arg values
	a.Silent = parseBoolArg(args, Silent, "running silently")
	a.SortYear = parseBoolArg(args, SortYear, "sorting years")
	a.SortMonth = parseBoolArg(args, SortMonth, "sorting months")
	a.SortGameType = parseBoolArg(args, SortGameType, "sorting game types")
	a.KeepPrefix = parseBoolArg(args, KeepPrefix, "keeping demo prefixes")
	a.RenameMap = parseBoolArg(args, RenameMap, "renaming with map name")
	a.RenameDuration = parseBoolArg(args, RenameDuration, "renaming with demo duration")
	a.SearchDirs = parseBoolArg(args, SearchDirs, "searching folders")
	a.Multithread = parseBoolArg(args, Multithread, "running on multiple threads")
	a.DateMajorDir = parseBoolArg(args, DateMajorDir, "using date-major directories")
	a.UseEditDate = parseBoolArg(args, UseEditDate, "using date that demo was last edited")
	a.TwelveHourTime = parseBoolArg(args, TwelveHourTime, "using twelve-hour time")
	a.ShowConVars = parseBoolArg(args, ShowConVars, "showing console variables from demos")

	// Get int arg values
	a.ZipOlderThan = uint8(parseIntArg(args, ZipOlderThan))
	if a.ZipOlderThan != 0 {
		fmt.Printf("Zipping demos older than: %d years\n", a.ZipOlderThan)
	} else {
		fmt.Println("Not zipping old demos")
	}

	a.CullBelow = uint8(parseIntArg(args, CullBelow))
	if a.CullBelow != 0 {
		fmt.Printf("Culling demos shorter than: %d seconds\n", a.CullBelow)
	} else {
		fmt.Println("Not culling short demos")
	}

	// Get snipe file
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, Snipe+"=") {
			a.Snipe = arg[len(Snipe)+1:]
			fmt.Println("Sniping file:", a.Snipe)
		}
	}

	// Get ignore words
	if i := slices.Index(args, IgnoreWords); i != -1 {
		a.IgnoreWords = append(a.IgnoreWords, args[i+1:]...)
		fmt.Println("Ignoring words:", a.IgnoreWords)
	}

	return a
}

/* Prompts user to press enter to exit the program, then exit when entered */
func enterToExit(silent bool) {
	if !silent {
		fmt.Println("Press enter to close.")
		fmt.Scanln()
	}
	os.Exit(0)
}

func main() {
	// Get commandline arguments
	args := getArgs()

	// Get demos
	demoList := demoio.GetDemos(args.IgnoreWords)
	demoListCount := len(demoList)
	fmt.Printf("Scanning %d demos...\n", demoListCount)

	// Do incrementing through demoList here
	for i := range demoList {
		// Get date and times from demo title
		if !args.UseEditDate {
			demoList[i].DateTime = demoio.GetDateTime(demoList[i])
		}

		// Get new names
		if args.RenameMap || args.RenameDuration {
			timeFormat := demoio.GetTimeFormat(args.TwelveHourTime)
			demoio.GetNewName(demoList[i], timeFormat, args.KeepPrefix, args.RenameMap, args.RenameDuration)
		}
	}

	// TODO: This is currently not ideal
	// Ideally, when we move demos we just check if a year has been seen before.
	// If not, then we handle the file and remember that we've seen that year.
	// If we are using DateMajorDir = false then we handle/remember on a per-gametype basis.
	years := demoio.GroupYears(demoList)

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
					c, m, t := demoio.GroupGameTypes(demos[cs:ce])
					scanned[i] = GameTypes{c, m, t}
				})
				chunkStart += chunkLength
				chunkEnd += chunkLength
			}
			scanners.Go(func() {
				c, m, t := demoio.GroupGameTypes(demos[chunkStart:])
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
			casualDemos, mvmDemos, tournamentDemos = demoio.GroupGameTypes(demos)
		}

		// Move files to corresponding directories
		dirPrefix := filepath.Join(strconv.Itoa(year)+"demos", "")
		var movers sync.WaitGroup
		// TODO Format wishDir = sprintf(majorDir/minorDir)
		movers.Go(func() { demoio.MoveToDirectory(filepath.Join(dirPrefix, "casual"), casualDemos) })
		movers.Go(func() { demoio.MoveToDirectory(filepath.Join(dirPrefix, "mvm"), mvmDemos) })
		movers.Go(func() { demoio.MoveToDirectory(filepath.Join(dirPrefix, "tournament"), tournamentDemos) })
		movers.Wait()
	}

	enterToExit(args.Silent)
}
