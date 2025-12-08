package main

import (
	"demo-ark/demoark/pkg/demoio"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Demo = demoio.Demo

// TODO: Add CullGameType
type Arguments struct {
	Silent         bool     // Run program without prompts
	SortYear       bool     // Group demos by year
	SortMonth      bool     // Group demos by month
	SortGameType   bool     // Group demos by game type
	KeepPrefix     bool     // Rename options won't overwrite a detected ds_prefix
	RenameMap      bool     // Rename the demo to contain the map name
	RenameDuration bool     // Rename demo to contain the duration of the demo
	SearchDirs     bool     // Search subdirectories within the current directory
	Multithread    bool     // Allow the use of multiple cores / threads
	DateMajorDir   bool     // True: year/month/gametype/demo.dem | False: gametype/year/month/demo.dem
	UseEditDate    bool     // Use the date that a demo was last edited rather than date in its file name
	TwelveHourTime bool     // True: 12hr | false: 24hr
	SetAsideCulled bool     // Set aside culled demos to a "culled" directory, rather than deleting them
	TwoStageCull   bool     // Will first set aside culled demos, then on a subsequent run delete previously set aside demos
	ShowConVars    bool     // Outputs console variables parsed from demo (mainly for debugging)
	ZipOlderThan   uint8    // Zip demos older than this many years
	CullBelow      uint16   // Number of seconds that demos below that duration will be deleted
	Snipe          string   // Snipe a specific file (exactly) to execute program on (mainly for debugging)
	IgnoreWords    []string // Ignore file / folder names containing string
}

/* Parses boolean arguments and returns the boolean value */
func parseBoolArg(args []string, keyword string, def bool, message string) bool {
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

	return def
}

/* Parses integer arguments and returns the integer value */
func parseIntArg(args []string, keyword string, def uint16) uint16 {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			value, err := strconv.Atoi(arg[len(keyword)+1:])
			if err != nil {
				log.Printf("Invalid argument value: %s\n", arg)
				enterToExit(false)
			}

			return uint16(value)
		}
	}

	return def
}

/* Gets argument values */
func getArgs() Arguments {
	const Silent = "silent"                 //
	const SortYear = "sortyear"             //
	const SortMonth = "sortmonth"           // TODO Should this be kept?
	const SortGameType = "sortgametype"     //
	const KeepPrefix = "keepprefix"         //
	const RenameMap = "renamemap"           //
	const RenameDuration = "renameduration" //
	const SearchDirs = "searchdirs"         // TODO
	const Multithread = "multithread"       // TODO: Partial
	const DateMajorDir = "datemajordir"     //
	const UseEditDate = "useeditdate"       //
	const TwelveHourTime = "twelvehourtime" // Should this be kept?
	const SetAsideCulled = "setasideculled" //
	const TwoStageCull = "twostagecull"     // TODO
	const ShowConVars = "showconvars"       //
	const ZipOlderThan = "zipolderthan"     // TODO
	const CullBelow = "cullbelow"           //
	const Snipe = "snipe"                   //
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
		SetAsideCulled: true,
		TwoStageCull:   false,
		ShowConVars:    false,
		ZipOlderThan:   1,
		CullBelow:      30,
		Snipe:          "",
		IgnoreWords:    []string{"reference"},
	}

	// Get bool arg values
	a.Silent = parseBoolArg(args, Silent, a.Silent, "running silently")
	a.SortYear = parseBoolArg(args, SortYear, a.SortYear, "sorting years")
	a.SortMonth = parseBoolArg(args, SortMonth, a.SortMonth, "sorting months")
	a.SortGameType = parseBoolArg(args, SortGameType, a.SortGameType, "sorting game types")
	a.KeepPrefix = parseBoolArg(args, KeepPrefix, a.KeepPrefix, "keeping demo prefixes")
	a.RenameMap = parseBoolArg(args, RenameMap, a.RenameMap, "renaming with map name")
	a.RenameDuration = parseBoolArg(args, RenameDuration, a.RenameDuration, "renaming with demo duration")
	a.SearchDirs = parseBoolArg(args, SearchDirs, a.SearchDirs, "searching folders")
	a.Multithread = parseBoolArg(args, Multithread, a.Multithread, "running on multiple threads")
	a.DateMajorDir = parseBoolArg(args, DateMajorDir, a.DateMajorDir, "using date-major directories")
	a.UseEditDate = parseBoolArg(args, UseEditDate, a.UseEditDate, "using date that demo was last edited")
	a.TwelveHourTime = parseBoolArg(args, TwelveHourTime, a.TwelveHourTime, "using twelve-hour time")
	a.SetAsideCulled = parseBoolArg(args, SetAsideCulled, a.SetAsideCulled, "setting aside culled demos")
	a.TwoStageCull = parseBoolArg(args, TwoStageCull, a.TwoStageCull, "Using two stage culling")

	// Get int arg values
	a.CullBelow = uint16(parseIntArg(args, CullBelow, a.CullBelow))
	if a.CullBelow != 0 {
		// Don't allow culling more than 5 minute demos
		if a.CullBelow > 300 {
			log.Printf("Invalid CullBelow value. Cannot cull demos longer than 5 minutes.\n")
			enterToExit(false)
		}

		fmt.Printf("Culling demos shorter than: %d seconds\n", a.CullBelow)
	} else {
		fmt.Println("Not culling short demos")
	}

	tempZipOlderThan := parseIntArg(args, ZipOlderThan, uint16(a.ZipOlderThan))
	if tempZipOlderThan <= 255 {
		a.ZipOlderThan = uint8(tempZipOlderThan)
	}
	if a.ZipOlderThan != 0 {
		fmt.Printf("Zipping demos older than: %d years\n", a.ZipOlderThan)
	} else {
		fmt.Println("Not zipping old demos")
	}

	// Get ignored words
	if i := slices.Index(args, IgnoreWords); i != -1 {
		a.IgnoreWords = append(a.IgnoreWords, args[i+1:]...)
		fmt.Println("Ignoring words:", a.IgnoreWords)
	}

	// Get arguments mainly used for debugging
	a.ShowConVars = parseBoolArg(args, ShowConVars, a.ShowConVars, "showing console variables from demos")

	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, Snipe+"=") {
			a.Snipe = arg[len(Snipe)+1:]
			fmt.Println("Sniping file:", a.Snipe)
		}
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
	var demoList []demoio.Demo
	var demoListCount int
	if len(args.Snipe) > 0 {
		demoList = demoio.SnipeDemo(args.Snipe)
		demoListCount = 1
	} else {
		demoList = demoio.GetDemos(args.SearchDirs, args.IgnoreWords)
		demoListCount = len(demoList)
		fmt.Printf("Scanning %d demos...\n", demoListCount)
	}

	// Cull short demos prior to parsing information from demos
	var culledDemos []Demo
	if args.CullBelow > 0 {
		culledDemos = demoio.CullShortDemos(&demoList, args.CullBelow)
	}

	// Do incrementing through demoList here
	for i := range demoList {
		// Get date and times from demo title
		if !args.UseEditDate {
			demoList[i].DateTime = demoio.GetDateTime(demoList[i])
		}

		// Get new names
		if args.RenameMap || args.RenameDuration {
			timeFormat := demoio.GetTimeFormat(args.TwelveHourTime)
			demoio.GetNewName(&demoList[i], timeFormat, args.KeepPrefix, args.RenameMap, args.RenameDuration)
		}
	}

	// Sort and move demos
	demoio.SortDemos(demoList, culledDemos, args.SortYear, args.SortGameType, args.DateMajorDir, args.SetAsideCulled, args.ShowConVars)

	// Report that we're done
	enterToExit(args.Silent)
}
