package main

import (
	"demo-ark/demoark/internal/configio"
	"demo-ark/demoark/internal/util"
	"demo-ark/demoark/pkg/demoio"
	"fmt"
)

type Demo = demoio.Demo

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
	args := configio.GetArgs()

	// Just create shortcut
	if args.CreateShortcut {
		configio.CreateConfiguredShortcut(args)
		util.EnterToExit(args.Silent)
	}

	// Launch TF2 while program runs
	if args.LaunchTF2 {
		configio.LaunchGame()
	}

	// Get demos
	var demoList []Demo
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
	demoio.SortDemos(demoList, culledDemos, args.SortYear, args.SortGameType, args.DateMajorDir, args.ShowConVars, args.CullMode)

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
	util.EnterToExit(args.Silent)
}
