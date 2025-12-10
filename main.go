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

type Arguments struct {
	Silent         bool     // Run program without prompts
	SortYear       bool     // Group demos by year
	SortGameType   bool     // Group demos by game type
	KeepPrefix     bool     // Rename options won't overwrite a detected ds_prefix
	RenameMap      bool     // Rename the demo to contain the map name
	RenameDuration bool     // Rename demo to contain the duration of the demo
	SearchDirs     bool     // Search subdirectories within the current directory
	Multithread    bool     // Allow the use of multiple cores / threads
	DateMajorDir   bool     // True: year/month/gametype/demo.dem | False: gametype/year/month/demo.dem
	SetAsideCulled bool     // Set aside culled demos to a "culled" directory, rather than deleting them
	TwoStageCull   bool     // Will first set aside culled demos, then on a subsequent run delete previously set aside demos
	ShowConVars    bool     // Outputs console variables parsed from demo (mainly for debugging)
	ZipOlderThan   uint8    // Zip demos older than this many years
	CullBelow      uint16   // Number of seconds that demos below that duration will be deleted
	CullGameTypes  string   // Cull specified gametypes (t = Tournament, c = Casual, m = MvM)
	Snipe          string   // Snipe a specific file (exactly) to execute program on (mainly for debugging)
	IgnoreWords    []string // Ignore file / folder names containing string
}

/* Keywords for command line arguments */
const Silent = "silent"                 //
const SortYear = "sortyear"             //
const SortGameType = "sortgametype"     //
const KeepPrefix = "keepprefix"         //
const RenameMap = "renamemap"           //
const RenameDuration = "renameduration" //
const SearchDirs = "searchdirs"         // TODO
const Multithread = "multithread"       // TODO
const DateMajorDir = "datemajordir"     //
const SetAsideCulled = "setasideculled" //
const TwoStageCull = "twostagecull"     // TODO
const ShowConVars = "showconvars"       //
const ZipOlderThan = "zipolderthan"     //
const CullBelow = "cullbelow"           //
const CullGameTypes = "cullgametypes"   //
const Snipe = "snipe"                   //
const IgnoreWords = "ignorewords"       //

/* Parses boolean arguments and returns the boolean value */
func parseBoolArg(args []string, keyword string, def bool, message string) (bool, bool) {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			switch arg[len(keyword)+1:] {
			case "1":
				fallthrough
			case "true":
				fmt.Println(strings.ToUpper(message[:1]) + message[1:])
				return true, true
			case "0":
				fallthrough
			case "false":
				fmt.Println("Not", message)
				return false, true
			default:
				log.Printf("Invalid argument value: %s\n", arg)
				enterToExit(false)
			}
		}
	}

	return def, false
}

/* Parses integer arguments and returns the integer value */
func parseIntArg(args []string, keyword string, def uint16) (uint16, bool) {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			value, err := strconv.Atoi(arg[len(keyword)+1:])
			if err != nil {
				log.Printf("Invalid argument value: %s\n", arg)
				enterToExit(false)
			}

			return uint16(value), true
		}
	}

	return def, false
}

/* Parses the boolean value from a user response to a given prompt */
func parseBoolPrompt(prompt string, message string) bool {
	for {
		// Prompt user
		fmt.Print(prompt, " [Y] / [N]: ")

		// Receive input
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user", err)
			continue
		}

		// Parse input
		if strings.EqualFold(input, "y") {
			fmt.Println(strings.ToUpper(message[:1]) + message[1:])
			return true
		} else if strings.EqualFold(input, "n") {
			fmt.Println("Not", message)
			return false
		}
	}
}

/* Parse the integer value from a user response to a given prompt */
func parseIntPrompt(prompt string, max int) uint16 {
	for {
		// Prompt user
		fmt.Printf("%s 0 (disabled) - %d: ", prompt, max)

		// Receive input
		input := -1
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user", err)
			continue
		}

		// Parse input
		if 0 <= input && input <= max {
			return uint16(input)
		}
	}
}

/* Prompts user for options not set by command line arguments */
func promptArgs(a *Arguments, cmdArgs map[string]bool) {
	// TODO: Sort this and add a "default" escape code
	// Ignore Silent, it can only be set by command line argument
	// Prompt for values not given by command line arguments
	if !cmdArgs[SortYear] {
		a.SortYear = parseBoolPrompt("Would you like to sort demos into folders by year?", "sorting years")
	}
	if !cmdArgs[SortGameType] {
		fmt.Println()
		a.SortGameType = parseBoolPrompt("Would you like to sort demos into folders by game type?", "sorting game types")
	}
	if !cmdArgs[KeepPrefix] {
		fmt.Println()
		a.KeepPrefix = parseBoolPrompt("Would you like to keep the current prefix (title before the date) of demos?", "keeping demo prefixes")
	}
	if !cmdArgs[RenameMap] {
		fmt.Println()
		a.RenameMap = parseBoolPrompt("Would you like to add the map name to a demo's name?", "renaming with map names")
	}
	if !cmdArgs[RenameDuration] {
		fmt.Println()
		a.RenameDuration = parseBoolPrompt("Would you like to add the duration to a demo's name?", "renaming with demo durations")
	}
	if !cmdArgs[SearchDirs] {
		fmt.Println()
		a.SearchDirs = parseBoolPrompt("Would you like to search subdirectories (folders) within the current directory?", "searching subdirectories")
	}
	if !cmdArgs[Multithread] {
		fmt.Println()
		a.Multithread = parseBoolPrompt("Would you like the program to utilize multiple CPU threads?", "running on multiple threads")
	}
	if !cmdArgs[DateMajorDir] {
		fmt.Println()
		a.DateMajorDir = parseBoolPrompt("Would you like the folder structure to prioritize sorting by year (year/gametype/demo.dem)?", "using date-major directories")
	}
	if !cmdArgs[SetAsideCulled] {
		fmt.Println()
		a.SetAsideCulled = parseBoolPrompt("Would you like to set aside culled demos to a \"culled\" directory (folder), rather than deleting them?", "setting aside culled demos")
	}
	if !cmdArgs[TwoStageCull] {
		fmt.Println()
		a.TwoStageCull = parseBoolPrompt("Would you like to require two program runs to delete culled demos?\nThe first would set aside culled demos and the next would delete demos.", "using two stage culling")
	}
	if !cmdArgs[ShowConVars] {
		fmt.Println()
		a.ShowConVars = parseBoolPrompt("Would you like to show parsed console variables when determining demos' game types?", "showing parsed console variables")
	}

	if !cmdArgs[ZipOlderThan] {
		fmt.Println()
		a.ZipOlderThan = uint8(parseIntPrompt("Enter the minimum age of a demo in years to compress to a .zip file.", 255))
		if a.ZipOlderThan == 0 {
			fmt.Println("Not zipping old demos")
		} else {
			fmt.Printf("Zipping demos older than: %d years\n", a.ZipOlderThan)
		}
	}
	if !cmdArgs[CullBelow] {
		fmt.Println()
		a.CullBelow = parseIntPrompt("Enter the minimum length of a demo in seconds to mark for culling.", 300)
		if a.CullBelow == 0 {
			fmt.Println("Not culling short demos")
		} else {
			fmt.Printf("Culling demos shorter than: %d seconds\n", a.CullBelow)
		}
	}

	// Get CullGameTypes key string
	for !cmdArgs[CullGameTypes] {
		fmt.Println()
		fmt.Println("Enter game types for demos you'd like to mark for culling.")
		fmt.Print("T for Tournament, C for Casual, and/or M for MvM: ")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		}

		// Validate value length
		if len(input) > 2 {
			log.Printf("Invalid input: %s. Will not allow culling of all demos.\n", input)
			continue
		}

		// Validate value contents
		input = strings.ToLower(input)
		if !(strings.ContainsRune(input, 't') || strings.ContainsRune(input, 'c') || strings.ContainsRune(input, 'm')) {
			log.Printf("Invalid input: %s. Example: \"CM\" would mark Casual and MvM demos for culling.\n", input)
			continue
		}

		a.CullGameTypes = input
		fmt.Println("Culling game types:", a.CullGameTypes)
		break
	}

	// Get Snipe string
	for !cmdArgs[Snipe] {
		fmt.Println()
		fmt.Print("Enter exact filename to selectively execute on.\nOr press [Enter] with no input to continue: ")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		}

		a.Snipe = input
		fmt.Println("Sniping file:", a.Snipe)
		break
	}

	// Get IgnoreWords strings
	prompted := false
	for !cmdArgs[IgnoreWords] {
		if !prompted {
			fmt.Print("\nEnter a word that should tell the program to ignore demos containing specified word.\nOr press [Enter] with no input to continue: ")
		}
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		} else if err != nil && strings.HasSuffix(err.Error(), "newline") {
			if len(a.IgnoreWords) != 0 {
				fmt.Println("Ignoring demos with titles containing:", a.IgnoreWords)
			}
			break
		}

		a.IgnoreWords = append(a.IgnoreWords, input)
	}
}

/* Gets argument values */
func getArgs() Arguments {

	// Convert args to lower case
	var args []string
	for _, arg := range os.Args[1:] {
		// Don't set Snipe's value to lower case!
		if strings.EqualFold(arg[:5], Snipe) {
			args = append(args, strings.ToLower(arg[:5])+arg[5:])
		}

		args = append(args, strings.ToLower(arg))
	}

	// Set defaults
	a := Arguments{
		Silent:         false,
		SortYear:       true,
		SortGameType:   true,
		KeepPrefix:     true,
		RenameMap:      false,
		RenameDuration: false,
		SearchDirs:     false,
		Multithread:    true,
		DateMajorDir:   true,
		SetAsideCulled: true,
		TwoStageCull:   false,
		ShowConVars:    false,
		ZipOlderThan:   1,
		CullBelow:      30,
		CullGameTypes:  "",
		Snipe:          "",
		IgnoreWords:    []string{"reference"},
	}
	cmdArgs := make(map[string]bool)

	// Get bool arg values
	a.Silent, cmdArgs[Silent] = parseBoolArg(args, Silent, a.Silent, "running silently")
	a.SortYear, cmdArgs[SortYear] = parseBoolArg(args, SortYear, a.SortYear, "sorting years")
	a.SortGameType, cmdArgs[SortGameType] = parseBoolArg(args, SortGameType, a.SortGameType, "sorting game types")
	a.KeepPrefix, cmdArgs[KeepPrefix] = parseBoolArg(args, KeepPrefix, a.KeepPrefix, "keeping demo prefixes")
	a.RenameMap, cmdArgs[RenameMap] = parseBoolArg(args, RenameMap, a.RenameMap, "renaming with map names")
	a.RenameDuration, cmdArgs[RenameDuration] = parseBoolArg(args, RenameDuration, a.RenameDuration, "renaming with demo durations")
	a.SearchDirs, cmdArgs[SearchDirs] = parseBoolArg(args, SearchDirs, a.SearchDirs, "searching subdirectories")
	a.Multithread, cmdArgs[Multithread] = parseBoolArg(args, Multithread, a.Multithread, "running on multiple threads")
	a.DateMajorDir, cmdArgs[DateMajorDir] = parseBoolArg(args, DateMajorDir, a.DateMajorDir, "using date-major directories")
	a.SetAsideCulled, cmdArgs[SetAsideCulled] = parseBoolArg(args, SetAsideCulled, a.SetAsideCulled, "setting aside culled demos")
	a.TwoStageCull, cmdArgs[TwoStageCull] = parseBoolArg(args, TwoStageCull, a.TwoStageCull, "using two stage culling")

	// Get CullBelow int
	a.CullBelow, cmdArgs[CullBelow] = parseIntArg(args, CullBelow, a.CullBelow)
	if a.CullBelow != 0 && cmdArgs[CullBelow] {
		// Don't allow culling more than 5 minute demos
		if a.CullBelow > 300 {
			log.Printf("Invalid CullBelow value. Cannot cull demos longer than 5 minutes.\n")
			enterToExit(false)
		}
		fmt.Printf("Culling demos shorter than: %d seconds\n", a.CullBelow)
	} else if cmdArgs[CullBelow] {
		fmt.Println("Not culling short demos")
	}

	// Get CullGameTypes string
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, CullGameTypes+"=") {
			gameTypes := arg[len(CullGameTypes)+1:]
			// Validate value length
			if len(gameTypes) > 2 {
				log.Printf("Invalid CullGameType value: %s. Will not allow culling of all demos.\nUsage: CullGameType=cm (t = Tournament, c = Casual, m = MvM).\n", gameTypes)
				enterToExit(false)
			} else if len(gameTypes) == 0 {
				log.Printf("Invalid CullGameType value. Need game type key.\nUsage: CullGameType=cm (t = Tournament, c = Casual, m = MvM).\n")
				enterToExit(false)
			}
			// Validate value contents
			if !(strings.ContainsRune(gameTypes, 't') || strings.ContainsRune(gameTypes, 'c') || strings.ContainsRune(gameTypes, 'm')) {
				log.Printf("Invalid CullGameType value: %s\nUsage: CullGameType=cm (t = Tournament, c = Casual, m = MvM).\n", gameTypes)
				enterToExit(false)
			}

			a.CullGameTypes = gameTypes
			cmdArgs[CullGameTypes] = true
			fmt.Println("Culling GameTypes:", a.CullGameTypes)
			break
		}
	}

	// Get ZipOlderThan int
	var tempZipOlderThan uint16
	tempZipOlderThan, cmdArgs[ZipOlderThan] = parseIntArg(args, ZipOlderThan, uint16(a.ZipOlderThan))
	if tempZipOlderThan <= 255 {
		a.ZipOlderThan = uint8(tempZipOlderThan)
	}
	if a.ZipOlderThan != 0 && cmdArgs[ZipOlderThan] {
		fmt.Printf("Zipping demos older than: %d years\n", a.ZipOlderThan)
	} else if cmdArgs[ZipOlderThan] {
		fmt.Println("Not zipping old demos")
	}

	// Get ignored words
	if i := slices.Index(args, IgnoreWords); i != -1 {
		a.IgnoreWords = append(a.IgnoreWords, args[i+1:]...)
		cmdArgs[IgnoreWords] = true
		fmt.Println("Ignoring words:", a.IgnoreWords)
	}

	// Get arguments mainly used for debugging
	a.ShowConVars, cmdArgs[ShowConVars] = parseBoolArg(args, ShowConVars, a.ShowConVars, "showing parsed console variables")

	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, Snipe+"=") {
			a.Snipe = arg[len(Snipe)+1:]
			cmdArgs[Snipe] = true
			fmt.Println("Sniping file:", a.Snipe)
			break
		}
	}

	// Prompt user for any non-argument settings
	if !a.Silent {
		promptArgs(&a, cmdArgs)
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
	// TODO: Profiling
	/*c, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer c.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(c); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}*/

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

	// Cull short demos prior to parsing more intensive information from demos
	var culledDemos []Demo
	if args.CullBelow > 0 {
		culledDemos = demoio.CullShortDemos(&demoList, args.CullBelow)
	}

	// Do incrementing through demoList here (Probably slightly slower but cleaner)
	for i := range demoList {
		// Get date and times from demo title
		demoList[i].DateTime = demoio.GetDateTime(demoList[i])

		// Get new names
		if args.RenameMap || args.RenameDuration {
			timeFormat := "15-04-05"
			demoio.GetNewName(&demoList[i], timeFormat, args.KeepPrefix, args.RenameMap, args.RenameDuration)
		}

		// Get game types if needed
		if args.SortGameType && args.CullGameTypes != "" {
			demoio.GetGameType(&demoList[i], args.ShowConVars)
		}
	}

	// Mark demos of specified gametypes for culling
	if args.CullGameTypes != "" {
		culledDemos = append(culledDemos, demoio.CullGameTypes(&demoList, args.CullGameTypes)...)
	}

	// Sort and move demos
	demoio.SortDemos(demoList, culledDemos, args.SortYear, args.SortGameType, args.DateMajorDir, args.SetAsideCulled, args.ShowConVars)

	// Zip folders older than specified number of years
	if args.ZipOlderThan > 0 {
		demoio.ZipOldDemos(args.ZipOlderThan)
	}

	// TODO: Profiling
	/*pprof.StopCPUProfile()
	m, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer m.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	// Lookup("allocs") creates a profile similar to go test -memprofile.
	// Alternatively, use Lookup("heap") for a profile
	// that has inuse_space as the default index.
	if err := pprof.Lookup("allocs").WriteTo(m, 0); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}*/

	// Report that we're done
	enterToExit(args.Silent)
}
