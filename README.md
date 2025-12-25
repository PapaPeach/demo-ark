# Demo Ark
The automated demo archiving program for TF2 to help save you from the flood of demos recorded by ds_enable.

I wrote this because I have always been irritated by the shortcomings of TF2's Demo Support implementation which registered Casual as Tournament matches. When RGL and other leagues forced this to be enabled I got tired of manually toggling `ds_enable` and didn't want 4 hour Hightower games to consume my SSDs.

Demo Ark supports the following game types: community competitive (like RGL), Casual, Valve competitive, MvM, and community servers (community game modes such as VSH or Zombie Infection are considered community servers).

Here's a [Quick Showcase Video](https://youtu.be/rFJxBitEfhQ).

# Installation
1. Download latest version of Demo Ark from the [Releases page](https://github.com/PapaPeach/demo-ark/releases).
2. Place **demoark.exe** where TF2 records demos to (this can be configured in-game with `ds_dir ...`).
3. I recommend starting with a subset of demos or using "set aside culled demos" (`CullMode=0`) while configuring.
4. Double click **demoark.exe** to run it. Your computer will likely warn you about running an executable from an unknown creator, you can run anyway.
5. Follow the prompts to configure demo sorting to your preference.

# Automated Running
To simplify keeping your demos organized, Demo Ark is built to enable various ways to automate running. The simplest is built directly into the program. Demo Ark can create a shortcut that will run with your desired options and can even launch the game while it runs. To enable this:
1. Place **demoark.exe** where TF2 records demos to (this can be configured in-game with `ds_dir ...`).
2. Run **demoark.exe** and select the options for how you'd like to organize your demos.
3. When asked if you'd like to create a shortcut (prompt 16 / 17), select **Yes**.
4. When asked if you'd like the shortcut to launch TF2 alongside Demo Ark, select **Yes**.
5. Answer the remaining prompts. When finished, Demo Ark will create a shortcut with the desired settings that you can place wherever you'd like.

If you would like to automate Demo Ark in some other method, like using a game library manager such as [PlayNite](https://github.com/JosefNemec/Playnite/) to run it after you close TF2, then Demo Ark will accept command line arguments to enable automation with the options listed below.
# Options
Demo Ark can be run with a slew of options, configured either via the prompts upon running the program, or via arguments used to launch the program through the command line or similar means.
- Arguments are should be separated by spaces, with each argument consisting of a keyword, equal sign, and value (no spaces).
- Capitalization is not important (except for `Snipe` filenames).
- Boolean values can be given as true/false or 1/0.
- The IgnoreWords argument is the exception to this formatting, it should be the last argument, with all desired ignored words being separated with spaces.

For example: `./demo-ark silent=true sortYear=1 ZipOlderThan=3 SNIPE=the_med.dem IgnoreWords untouchable secret`  
| Key Word        |     Possible Values      | Default Value | Description |
|-----------------|:------------------------:|:-------------:|-------------|
| Silent          |   true (1) / false (0)   |     false     | Run the program without prompts |
| SortYear        |   true (1) / false (0)   |     true      | Group demos into folders by year |
| SortGameType    |   true (1) / false (0)   |     true      | Group demos into folders by gametype |
| DateMajorDir    |   true (1) / false (0)   |     true      | **True:** year/month/gametype/demo.dem<br>**False:** gametype/year/month/demo.dem |
| RenameMap       |   true (1) / false (0)   |     false     | Rename the demo to contain the map name |
| RenameDuration  |   true (1) / false (0)   |     false     | Rename demo to contain the duration of the demo |
| KeepPrefix      |   true (1) / false (0)   |     true      | Rename options won't overwrite a detected ds_prefix |
| SearchDirs      |   true (1) / false (0)   |     false     | Search subdirectories within the current directory |
| CreateShortcut  |   true (1) / false (0)   |     false     | **Windows / Linux:** Create a shortcut to run program with the selected options<br>**Mac:** Print a copyable path to the program with selected options |
| LaunchTF2       |   true (1) / false (0)   |     false     | Launch TF2 while program runs |
| CullEventTxts   |   true (1) / false (0)   |     false     | Cull all _events.txt files made by `ds_log 1`<br>**Note: Empty files will be culled regardless of setting** |
| CullEventJsons  |   true (1) / false (0)   |     false     | Cull all _.json event files made by `ds_log 1`<br>**Note: Empty files or files with no corresponding demo will be culled regardless of setting** |
| CullScreenshots |   true (1) / false (0)   |     false     | Cull all _.tga screenshot made by `ds_screens 1`<br>**Note: Screenshots with no corresponding demo will be culled regardless of setting** |
| CullMode        |          0 - 2           |       1       | **0:** Set aside culled demos to "culled" folder<br>**1:** Delete previously set aside demos, then set aside next batch of culled demos to be deleted on future Demo Ark runs<br>**2:** Delete demos immediately |
| CullGameTypes   |        c, q, m, v        |      "c"      | Cull demos of a specific game types: **C**asual, **Q**uickPlay (Community), **M**vM, **V**alve Competitive<br>**Example:** `cullgametypes=cm` Will cull **C**asual and **M**vM |
| CullBelow       |    0 (disabled) - 300    |      30       | Number of seconds that demos below that duration will be deleted |
| ZipOlderThan    |    0 (disabled) - 255    |       1       | Zip demos older than this many years (will wait until Spring season starts to zip previous year)<br>**Note: Compression takes about 0.5s per unsorted demo** |
| IgnoreWords     | Any words after key word |   "ignore"    | Ignore file / folder names containing string<br>**Note: This must be the last argument (other than its keywords)** |
| Snipe           | Any continuous filename  |      ""       | Snipe a specific file (exactly) to execute program on (mainly for debugging) |
| ShowConVars     |   true (1) / false (0)   |     false     | Outputs console variables parsed from demo (mainly for debugging) |

# Potential Options
I tried to keep the options limited to things most people would find useful to keep customization approachable and maintainable. Unfortunately, I can't please everyone, but I think the program covers an overwhelming majority of use cases.

I can't add one-off customization options for individuals, if I receive enough feedback for features via the appropriate channels, such as the TF2Utilities Feedback channel my [HUD / Project Discord](https://discord.gg/HyZRVtp), I will do my best to add them.
| Key Word       |        Status        | Description |
|----------------|:--------------------:|-------------|
| Multithread    | Planned, see below | Allow the use of multiple cores / threads<br> **0:** Off **1:** All but 1 core **2:** Half cores **3:** Quarter cores... |
| SortMonth      |  Removed, too niche  | Group demos into folders by month |
| UseEditDate    |  Removed, too niche  | Use the date that a demo was last edited rather than date in its file name |
| TwelveHourTime | Removed, too finicky | **True:** 12hr<br>**False:** 24hr |

**Note on Multithreading:** This was originally a planned feature at launch. But after farther consideration this was postponed as it was determined to have limited usefulness and require more testing.

The bulk of processing time for Demo Ark is spent on disk read / write operations that would see no benefit from CPU parallelization. Furthermore, multithreading has an initial overhead that would only break even for bulk sorting of ~10 demos per core, which would likely only be the initial run of Demo Ark.  
On the frequent smaller operations intended with the automated functionality, the initial overhead of multithreading would actually *increase* sorting times, though this would be recognized by the program and disabled.

It is still a planned feature, but that is why it is not available on release.

# Supported Operating Systems
| Operating System | Support |
|------------------|---------|
| Windows | ✅ Full |
| Linux<br>(SteamOS) | ✅ Full |
| Mac | 🆗 Requires building from source<br>Only semi-automated shortcut creation |

# How's It Work?
Demos pre-date the ability to conveniently share videos, so instead they are essentially instructions for an offline server to replay the game exactly as the original live match you recorded. That includes server settings which can be used to differentiate between the settings used for Casual, Tournament, and MvM servers.  

The rest of the program is simply parsing the bit-buffered contents of demos and managing selected options.

# Is This A Virus?
### Nope.
Your anti-virus will likely warn you about running executables made by unknown creators, which is its job. If you know how, I invite you to review my code, build it from source, and even provide feedback.

If the project takes off I can look into getting Demo Ark officially approved by Microsoft so Windows Security won't show this warning. Though to maintain Microsoft recognition I'd have to resubmit Demo Ark every update and wait for their approval, or pay $300-$700 annually for Microsoft to guarantee my program would be reviewed within 24 hours of submission. This also wouldn't apply to third-party anti-virus providers.

# Thanks
**[NeKzor's Portal 2 Demo Documentation](https://dem.nekz.me/)** - The most comprehensive documentation on demos I've ever seen, for any Source game.  
**[Klauspost's Compress Go Package](https://github.com/klauspost/compress)** - A drop-in 15% - 25% optimization to the Go Standard Library Zip package.  
**[Pektezol's BitReader Go Package](https://github.com/pektezol/bitreader)** - An accessible Go package for parsing bit-buffered demo contents.  
**[Jxeng's Shortcut Go Package](https://github.com/jxeng/shortcut)** - A Go package that greatly simplifies Windows shortcut creation.  
**[Demostf's Demo Parser](https://codeberg.org/demostf/parser)** - My reference for TF2 specific bit-values, packet, and message specs.  
**Ward** - Providing feedback and testing predecessor programs and pre-release versions of Demo Ark.
