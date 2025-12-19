package configio

import (
	"demo-ark/demoark/internal/util"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

type Arguments struct {
	Silent         bool     // Run program without prompts
	SortYear       bool     // Group demos by year
	SortGameType   bool     // Group demos by game type
	DateMajorDir   bool     // True: year/month/gametype/demo.dem | False: gametype/year/month/demo.dem
	KeepPrefix     bool     // Rename options won't overwrite a detected ds_prefix
	RenameMap      bool     // Rename the demo to contain the map name
	RenameDuration bool     // Rename demo to contain the duration of the demo
	SearchDirs     bool     // Search subdirectories within the current directory
	Multithread    bool     // Allow the use of multiple cores / threads
	CreateShortcut bool     // Create a shortcut with the currently selected options
	LaunchTF2      bool     // Launch TF2 alongside the program's execution
	ShowConVars    bool     // Outputs console variables parsed from demo (mainly for debugging)
	ZipOlderThan   uint8    // Zip demos older than this many years
	CullMode       uint8    // 0: Set aside culled demos to a "culled" directory 1: Delete previously set aside demos, then set aside culled demos 2: Delete culled demos
	CullBelow      uint16   // Number of seconds that demos below that duration will be deleted
	CullGameTypes  string   // Cull specified gametypes (t = Tournament, c = Casual, m = MvM)
	Snipe          string   // Snipe a specific file (exactly) to execute program on (mainly for debugging)
	IgnoreWords    []string // Ignore file / folder names containing string
}

// Default option values
var def = Arguments{
	Silent:         false,
	SortYear:       true,
	SortGameType:   true,
	DateMajorDir:   true,
	KeepPrefix:     true,
	RenameMap:      false,
	RenameDuration: false,
	SearchDirs:     false,
	Multithread:    true,
	CreateShortcut: false,
	LaunchTF2:      false,
	ShowConVars:    false,
	ZipOlderThan:   1,
	CullMode:       0,
	CullBelow:      30,
	CullGameTypes:  "",
	Snipe:          "",
	IgnoreWords:    []string{"ignore"},
}

/* Keywords for command line arguments. */
const Silent = "silent"
const SortYear = "sortyear"
const SortGameType = "sortgametype"
const DateMajorDir = "datemajordir"
const KeepPrefix = "keepprefix"
const RenameMap = "renamemap"
const RenameDuration = "renameduration"
const SearchDirs = "searchdirs"
const Multithread = "multithread" // TODO
const CreateShortcut = "createshortcut"
const LaunchTF2 = "launchtf2"
const ShowConVars = "showconvars"
const ZipOlderThan = "zipolderthan"
const CullMode = "cullmode"
const CullBelow = "cullbelow"
const CullGameTypes = "cullgametypes"
const Snipe = "snipe"
const IgnoreWords = "ignorewords"

/* Parses boolean arguments and returns the boolean value. */
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
				util.EnterToExit(false)
			}
		}
	}

	return def, false
}

/* Parses integer arguments and returns the integer value. */
func parseIntArg(args []string, keyword string, def uint16) (uint16, bool) {
	for _, arg := range args {
		// Locate keyword=...
		if strings.HasPrefix(arg, keyword+"=") {
			value, err := strconv.ParseUint(arg[len(keyword)+1:], 10, 0)
			if err != nil {
				log.Printf("Invalid argument value: %s\n", arg)
				util.EnterToExit(false)
			}

			return uint16(value), true
		}
	}

	return def, false
}

/* Parses the boolean value from a user response to a given prompt. */
func parseBoolPrompt(prompt string, message string, def bool) (bool, bool) {
	for {
		// Prompt user
		fmt.Print(prompt, "\n[Y]es / [N]o / [D]efault / [B]ack: ")

		// Receive input
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		}

		// Parse input
		input = strings.ToLower(input[:1])
		switch input {
		case "y": // Yes
			fmt.Println(strings.ToUpper(message[:1]) + message[1:])
			return true, false
		case "n": // No
			fmt.Println("Not", message)
			return false, false
		case "d": // Default
			fmt.Println("Using default setting")
			return def, false
		case "b": // Back
			return def, true
		}
	}
}

/* Parse the integer value from a user response to a given prompt. */
func parseIntPrompt(prompt string, max uint64, def uint16) (uint16, bool) {
	for {
		// Prompt user
		fmt.Printf("%s\n0 (disabled) - %d / [D]efault / [B]ack: ", prompt, max)

		// Receive input
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		}

		// Parse text input
		input = strings.ToLower(input[:1])
		switch input {
		case "d": // Default
			fmt.Println("Using default setting")
			return def, false
		case "b": // Back
			return def, true
		}

		// Parse input
		inputInt, err := strconv.ParseUint(input, 10, 0)
		if err != nil && !strings.HasSuffix(err.Error(), "invalid syntax") { // Genuine error
			log.Println("Error parsing response to integer prompt from user:", err)
		}
		if inputInt <= max {
			return uint16(inputInt), false
		}
	}
}

/* Prompts user for options not set by command line arguments. */
func promptArgs(a *Arguments, cmdArgs map[string]bool) {
	defer fmt.Println()
prompt0:
	// Explain prompts
	fmt.Println("Some options were not set by the program launch arguments. There may be up to 16 options to set.")
	for {
		fmt.Print("Would you like to skip setting options and run with remaining options set to defaults?\nPress [Enter] to view remaining options or type \"skip\" to skip: ")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil {
			// View remaining options
			if strings.HasSuffix(err.Error(), "newline") {
				break
			}
			// Genuine error
			log.Println("Error getting input from user:", err)
			continue
		}

		if strings.EqualFold(input, "skip") {
			fmt.Println("Using default options")
			return
		}
	}
	fmt.Println("Prompting for remaining options, some prompts may be skipped based on previous answers.")

	// Ignore Silent, it can only be set by command line argument
	// Sort options
prompt1:
	if !cmdArgs[SortYear] {
		fmt.Println()
		back := false
		a.SortYear, back = parseBoolPrompt("(1 / 17) Would you like to sort demos into folders by year?", "sorting years", def.SortYear)
		if back {
			goto prompt0
		}
	}
prompt2:
	if !cmdArgs[SortGameType] {
		fmt.Println()
		back := false
		a.SortGameType, back = parseBoolPrompt("(2 / 17) Would you like to sort demos into folders by game type?", "sorting game types", def.SortGameType)
		if back { // Handle back command
			goto prompt1
		}
	}
prompt3:
	if !cmdArgs[DateMajorDir] && (a.SortYear || a.SortGameType) { // Only ask if we're sorting into folders at all
		fmt.Println()
		back := false
		a.DateMajorDir, back = parseBoolPrompt("(3 / 17) Would you like the folder structure to prioritize sorting by year (year/gametype/demo.dem)?", "using date-major directories", def.DateMajorDir)
		if back { // Handle back command
			goto prompt2
		}
	}

	// Rename options
prompt4:
	if !cmdArgs[RenameMap] {
		fmt.Println()
		back := false
		a.RenameMap, back = parseBoolPrompt("(4 / 17) Would you like to add the map name to a demo's name?", "renaming with map names", def.RenameMap)
		if back { // Handle back command
			if a.SortYear || a.SortGameType { // Special case for skip logic
				goto prompt3
			}
			goto prompt2
		}
	}
prompt5:
	if !cmdArgs[RenameDuration] {
		fmt.Println()
		back := false
		a.RenameDuration, back = parseBoolPrompt("(5 / 17) Would you like to add the duration to a demo's name?", "renaming with demo durations", def.RenameDuration)
		if back { // Handle back command
			goto prompt4
		}
	}
prompt6:
	if !cmdArgs[KeepPrefix] && (a.RenameMap || a.RenameDuration) { // Only ask if we are renaming at all
		fmt.Println()
		back := false
		a.KeepPrefix, back = parseBoolPrompt("(6 / 17) Would you like to keep the current prefix (title before the date) of demos?\nMap names and demo duration would be added in addition to the prefix.", "keeping demo prefixes", def.KeepPrefix)
		if back { // Handle back command
			goto prompt5
		}
	}

	// Culling options
prompt7:
	if !cmdArgs[CullBelow] {
		fmt.Println()
		back := false
		a.CullBelow, back = parseIntPrompt("(7 / 17) Enter the minimum length of a demo in seconds to mark for culling.", 300, def.CullBelow)
		if back { // Handle back command
			if a.RenameMap || a.RenameDuration { // Special case for skip logic
				goto prompt6
			}
			goto prompt5
		}

		// Handle option value
		if a.CullBelow == 0 {
			fmt.Println("Not culling short demos")
		} else {
			fmt.Printf("Culling demos shorter than: %d seconds\n", a.CullBelow)
		}
	}

	// Get CullGameTypes key string
prompt8:
	for !cmdArgs[CullGameTypes] {
		fmt.Println()
		fmt.Println("(8 / 17) Enter game types for demos you'd like to mark for culling.")
		fmt.Print("Presse [Enter] with no input to skip / [T]ournament / [C]asual / [Q]uickPlay (community) / [M]vM / [V]alve Competitive / [D]efault / [B]ack:")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil {
			// Skip
			if strings.HasSuffix(err.Error(), "newline") {
				fmt.Println("Not culling any game types")
				break
			}
			// Genuine error
			log.Println("Error getting input from user:", err)
			continue
		}

		// Handle Default or Back command
		switch strings.ToLower(input[:1]) {
		case "d":
			fmt.Println("Using default value")
			a.CullGameTypes = def.CullGameTypes
			goto prompt9
		case "b":
			goto prompt7
		}

		// Validate value length
		if len(input) > 4 {
			log.Printf("Invalid input: %s. Will not allow culling of all demos.\n", input)
			continue
		}

		// Validate value contents
		input = strings.ToLower(input)
		if !(strings.ContainsRune(input, 't') ||
			strings.ContainsRune(input, 'c') ||
			strings.ContainsRune(input, 'q') ||
			strings.ContainsRune(input, 'm') ||
			strings.ContainsRune(input, 'v')) {
			log.Printf("Invalid input: %s. Example: \"CM\" would mark Casual and MvM demos for culling.\n", input)
			continue
		}

		a.CullGameTypes = input
		fmt.Println("Culling game types:", a.CullGameTypes)
		break
	}

prompt9:
	// Only ask if we're culling at all
	if !cmdArgs[CullMode] && (a.CullBelow != 0 || len(a.CullGameTypes) != 0) {
		fmt.Println()
		for {
			// Prompt user
			fmt.Print("(9 / 17) What would you like to do with culled demos?\n[0] Set aside to \"culled\" folder / [1] Set aside to be deleted on the next Demo Ark run / [2] Delete immediately / [D]efault / [B]ack: ")

			// Receive input
			input := ""
			_, err := fmt.Scanln(&input)
			if err != nil && !strings.HasSuffix(err.Error(), "newline") {
				log.Println("Error getting input from user:", err)
				continue
			}

			// Parse text input
			input = strings.ToLower(input[:1])
			switch input {
			case "d": // Default
				fmt.Println("Using default setting")
				a.CullMode = def.CullMode
				goto prompt10
			case "b": // Back
				goto prompt8
			}

			// Parse input
			inputInt, err := strconv.ParseUint(input, 10, 0)
			if err != nil {
				if strings.HasSuffix(err.Error(), "invalid syntax") { // Bad input
					continue
				}
				// Genuine error
				log.Println("Error parsing response to integer prompt from user:", err)
				util.EnterToExit(false)
			}
			if inputInt <= 2 {
				a.CullMode = uint8(inputInt)
				break
			}
		}

		// Handle option values
		switch a.CullMode {
		case 0: // Set aside
			fmt.Println("Culled demos will be set aside")
		case 1: // Two-stage delete
			fmt.Println("Culled demos will be set aside and deleted on future Demo Ark runs")
		case 2: // Simple delete
			fmt.Println("Culled demos will be deleted")
		}
	}

	// Get ZipOlderThan int
prompt10:
	if !cmdArgs[ZipOlderThan] {
		fmt.Println()
		tempZipOlderThan, back := parseIntPrompt("(10 / 17) Enter the minimum age of a demo in years to compress to a .zip file.", 255, uint16(def.ZipOlderThan))
		if back { // Handle back command
			if a.CullBelow != 0 || len(a.CullGameTypes) != 0 { // Special case for skip logic
				goto prompt9
			}
			goto prompt8
		}

		// Handle option value
		a.ZipOlderThan = uint8(tempZipOlderThan)
		if a.ZipOlderThan == 0 {
			fmt.Println("Not zipping old demos")
		} else {
			fmt.Printf("Zipping demos older than: %d years\n", a.ZipOlderThan)
		}
	}

	// Get IgnoreWords strings
prompt11:
	prompted := false
	for !cmdArgs[IgnoreWords] {
		if !prompted {
			fmt.Print("\n(11 / 17) Enter a word that should tell the program to ignore demos containing specified word.\nPress [Enter] with no input to skip / [D]efault / [B]ack: ")
		} else {
			fmt.Print("Press [Enter] with no input to continue: ")
		}
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") { // Genuine error
			log.Println("Error getting input from user:", err)
			continue
		} else if err != nil && strings.HasSuffix(err.Error(), "newline") { // Done
			if len(a.IgnoreWords) != 0 {
				fmt.Println("Ignoring demos with titles containing:", a.IgnoreWords)
			}
			break
		}

		// Handle Default or Back command
		if !prompted {
			switch strings.ToLower(input) {
			case "d":
				fmt.Println("Using default value")
				a.IgnoreWords = def.IgnoreWords
				goto prompt12
			case "b":
				goto prompt10
			}

			prompted = true
		}

		// Enforce minimum length
		if len(input) < 2 {
			fmt.Println("Ignored words must be at least 2 letters long")
			continue
		}

		a.IgnoreWords = append(a.IgnoreWords, input)
	}

	// Misc
prompt12:
	if !cmdArgs[SearchDirs] {
		fmt.Println()
		back := false
		a.SearchDirs, back = parseBoolPrompt("(12 / 17) Would you like to search subdirectories (folders) within the current directory?", "searching subdirectories", def.SearchDirs)
		if back { // Handle back command
			goto prompt11
		}
	}
prompt13:
	if !cmdArgs[Multithread] { // This is in a dumb spot
		fmt.Println()
		back := false
		a.Multithread, back = parseBoolPrompt("(13 / 17) Would you like the program to utilize multiple CPU threads?", "running on multiple threads", def.Multithread)
		if back { // Handle back command
			goto prompt12
		}
	}

	// Get shortcut options
prompt14:
	if !cmdArgs[CreateShortcut] {
		fmt.Println()
		back := false
		a.CreateShortcut, back = parseBoolPrompt("(14 / 17) Would you like to create a shortcut to launch Demo Ark with selected options?", "creating a shortcut to run program with selected options", def.CreateShortcut)
		if back { // Handle back command
			goto prompt13
		}
	}
prompt15:
	if !cmdArgs[LaunchTF2] && a.CreateShortcut {
		fmt.Println()
		back := false
		a.LaunchTF2, back = parseBoolPrompt("(15 / 17) Would you like the shortcut to launch TF2 alongside Demo Ark?", "launching TF2 while program runs", def.LaunchTF2)
		if back { // Handle back command
			goto prompt14
		}
	}

	// Debug stuff: ShowConVars and Snipe
	for {
		fmt.Print("Would you like to skip the remaining options, mainly used for debugging the program?\nPress [Enter] to view remaining options or type \"skip\" to skip: ")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil {
			// View remaining options
			if strings.HasSuffix(err.Error(), "newline") {
				break
			}
			// Genuine error
			log.Println("Error getting input from user:", err)
			continue
		}

		if strings.EqualFold(input, "skip") {
			fmt.Println("Skipping debug options")
			return
		}
	}

prompt16:
	if !cmdArgs[ShowConVars] {
		fmt.Println()
		back := false
		a.ShowConVars, back = parseBoolPrompt("(16 / 17) Would you like to show parsed console variables when determining demos' game types?", "showing parsed console variables", def.ShowConVars)
		if back { // Handle back command
			if a.CreateShortcut { // Special case for skip logic
				goto prompt15
			}
			goto prompt14
		}
	}

	// Get Snipe string
	for !cmdArgs[Snipe] {
		fmt.Println()
		fmt.Print("(17 / 17) Enter exact filename to selectively execute on.\nPress [Enter] with no input to skip / [D]efault / [B]ack: ")
		input := ""
		_, err := fmt.Scanln(&input)
		if err != nil && !strings.HasSuffix(err.Error(), "newline") {
			log.Println("Error getting input from user:", err)
			continue
		}

		// Handle Default or Back command
		switch strings.ToLower(input) {
		case "d":
			fmt.Println("Using default value")
			a.Snipe = def.Snipe
			return
		case "b":
			goto prompt16
		}

		// Enforce minimum length
		if len(input) < 2 {
			fmt.Println("Filenames must be at least 2 letters long")
			continue
		}

		a.Snipe = input
		fmt.Println("Sniping file:", a.Snipe)
		break
	}
}

/* Launches TF2 via Steam api to maintain user's launch options. */
func LaunchGame() {
	// Get correct launch command for given OS
	url := "steam://rungameid/440"
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin": // Mac
		cmd = exec.Command("open", url)
	default:
		log.Println("Detected unsupported operating system.")
		util.EnterToExit(false)
	}

	// Run command and continue with program execution
	err := cmd.Start()
	if err != nil {
		log.Println("Error launching TF2:", err)
		util.EnterToExit(false)
	}
}

/* Gets argument values. */
func GetArgs() Arguments {

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
	a := def
	cmdArgs := make(map[string]bool)

	// Get bool arg values
	a.Silent, cmdArgs[Silent] = parseBoolArg(args, Silent, a.Silent, "running silently")
	a.SortYear, cmdArgs[SortYear] = parseBoolArg(args, SortYear, a.SortYear, "sorting years")
	a.SortGameType, cmdArgs[SortGameType] = parseBoolArg(args, SortGameType, a.SortGameType, "sorting game types")
	a.DateMajorDir, cmdArgs[DateMajorDir] = parseBoolArg(args, DateMajorDir, a.DateMajorDir, "using date-major directories")
	a.KeepPrefix, cmdArgs[KeepPrefix] = parseBoolArg(args, KeepPrefix, a.KeepPrefix, "keeping demo prefixes")
	a.RenameMap, cmdArgs[RenameMap] = parseBoolArg(args, RenameMap, a.RenameMap, "renaming with map names")
	a.RenameDuration, cmdArgs[RenameDuration] = parseBoolArg(args, RenameDuration, a.RenameDuration, "renaming with demo durations")
	a.SearchDirs, cmdArgs[SearchDirs] = parseBoolArg(args, SearchDirs, a.SearchDirs, "searching subdirectories")
	a.Multithread, cmdArgs[Multithread] = parseBoolArg(args, Multithread, a.Multithread, "running on multiple threads")

	// Get CullMode int
	var tempCullMode uint16
	tempCullMode, cmdArgs[CullMode] = parseIntArg(args, CullMode, uint16(a.CullMode))
	// Enforce bounds
	if tempCullMode <= 2 {
		a.CullMode = uint8(tempCullMode)
	} else { // Above upper bound
		log.Println("Invalid CullMode value. Allowed range: 0 to 2.")
		util.EnterToExit(false)
	}
	if cmdArgs[CullMode] {
		switch a.CullMode {
		case 0: // Set aside
			fmt.Println("Culled demos will be set aside")
		case 1: // Two-stage delete
			fmt.Println("Culled demos will be set aside and deleted on future Demo Ark runs")
		case 2: // Simple delete
			fmt.Println("Culled demos will be deleted")
		}
	}

	// Get CullBelow int
	a.CullBelow, cmdArgs[CullBelow] = parseIntArg(args, CullBelow, a.CullBelow)
	if a.CullBelow != 0 && cmdArgs[CullBelow] {
		// Don't allow culling more than 5 minute demos
		if a.CullBelow > 300 {
			log.Println("Invalid CullBelow value. Cannot cull demos longer than 5 minutes.")
			util.EnterToExit(false)
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
			if len(gameTypes) > 4 {
				log.Printf("Invalid CullGameType value: %s. Will not allow culling of all demos.\nUsage: CullGameType=cm (t = Tournament, c = Casual, q = QuickPlay / Community, m = MvM, v = Valve Competitive).\n", gameTypes)
				util.EnterToExit(false)
			} else if len(gameTypes) == 0 {
				log.Printf("Invalid CullGameType value. Need game type key.\nUsage: CullGameType=cm (t = Tournament, c = Casual, q = QuickPlay / Community, m = MvM, v = Valve Competitive).\n")
				util.EnterToExit(false)
			}
			// Validate value contents
			if !(strings.ContainsRune(gameTypes, 't') ||
				strings.ContainsRune(gameTypes, 'c') ||
				strings.ContainsRune(gameTypes, 'q') ||
				strings.ContainsRune(gameTypes, 'm') ||
				strings.ContainsRune(gameTypes, 'v')) {
				log.Printf("Invalid CullGameType value: %s\nUsage: CullGameType=cm (t = Tournament, c = Casual, q = QuickPlay / Community, m = MvM, v = Valve Competitive).\n", gameTypes)
				util.EnterToExit(false)
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
	// Enforce bounds
	if tempZipOlderThan <= 255 {
		a.ZipOlderThan = uint8(tempZipOlderThan)
	} else { // Above upper bound
		log.Println("Invalid ZipOlderThan value. Allowed range: 0 to 255.")
		util.EnterToExit(false)
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

	a.LaunchTF2, cmdArgs[LaunchTF2] = parseBoolArg(args, LaunchTF2, a.LaunchTF2, "launching TF2 while program runs")
	a.CreateShortcut, cmdArgs[CreateShortcut] = parseBoolArg(args, CreateShortcut, a.CreateShortcut, "creating a shortcut to run program with selected options")

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
