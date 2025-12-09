# Demo Ark
The automated demo archiving program for TF2 to help save you from the flood of demos recorded by ds_enable.  
I wrote this because I have always been irritated by the shortcomings of TF2's Demo Support implementation which registered Casual as Tournament matches. When RGL and other leagues forced this to be enabled I got tired of manually toggling `ds_enable` and didn't want 4 hour Hightower games to consume my SSDs.

# How's It Work?
Demos pre-date the ability to conveniently share videos, so instead they are essentially instructions for an offline server to replay the exact game as the original live game you recorded. That includes server settings which can be used to differentiate between the settings used for Casual, Tournament, and MvM servers.  
The rest of the program is simply parsing the bit-buffered contents of demos and managing selected options.

# Options
Demo Ark can be run with a slew of options, configured either via the prompts upon running the program, or via arguments used to launch the program through the command line or similar means.  
Arguments are should be separated by spaces, with each argument consisting of a keyword, equal sign, and value (no spaces). Capitalization is not important. Boolean values can be given as true/false or 1/0.  
The IgnoreWords argument is the exception to this formatting, it should be the last argument, with all desired ignored words being separated with spaces.  
For example:  
`./demo-ark silent=true sortYear=1 ZipOlderThan=3 SNIPE=the_med.dem IgnoreWords untouchable secret`  
| Key Word       |     Possible Values      | Default Value | Description                                                                                                        |
|----------------|:------------------------:|:-------------:|--------------------------------------------------------------------------------------------------------------------|
| Silent         |   true (1) / false (0)   |     false     | Run the program without prompts                                                                                    |
| SortYear       |   true (1) / false (0)   |     true      | Group demos into folders by year                                                                                   |
| SortGameType   |   true (1) / false (0)   |     true      | Group demos into folders by gametype                                                                               |
| KeepPrefix     |   true (1) / false (0)   |     true      | Rename options won't overwrite a detected ds_prefix                                                                |
| RenameMap      |   true (1) / false (0)   |     false     | Rename the demo to contain the map name                                                                            |
| RenameDuration |   true (1) / false (0)   |     false     | Rename demo to contain the duration of the demo                                                                    |
| SearchDirs     |   true (1) / false (0)   |     false     | Search subdirectories within the current directory                                                                 |
| Multithread    |   true (1) / false (0)   |     true      | Allow the use of multiple cores / threads                                                                          |
| DateMajorDir   |   true (1) / false (0)   |     true      | **True:** year/month/gametype/demo.dem<br>**False:** gametype/year/month/demo.dem                                  |
| UseEditDate    |   true (1) / false (0)   |     false     | Use the date that a demo was last edited rather than date in its file name                                         |
| SetAsideCulled |   true (1) / false (0)   |     true      | Set aside culled demos to a "culled" directory, rather than deleting them                                          |
| ShowConVars    |   true (1) / false (0)   |     false     | Outputs console variables parsed from demo (mainly for debugging)                                                  |
| TwoStageCull   |   true (1) / false (0)   |     false     | Will first set aside culled demos, then on a subsequent run delete previously set aside demos                      |
| CullBelow      |    0 (disabled) - 300    |      30       | Number of seconds that demos below that duration will be deleted                                                   |
| ZipOlderThan   |    0 (disabled) - 255    |       1       | Zip demos older than this many years<br>**Note: Compression takes about 0.4s per demo**                            |
| Snipe          | Any continuous filename  |      ""       | Snipe a specific file (exactly) to execute program on (mainly for debugging)                                       |
| IgnoreWords    | Any words after key word |  "reference"  | Ignore file / folder names containing string<br>**Note: This must be the last argument (other than its keywords)** |

# Potential Options
I tried to keep the options limited to things most people would find useful to keep customization approachable and maintainable. Unfortunately, I can't please everyone, but I think the program covers an overwhelming majority of use cases.  
I can't add one-off customization options for individuals, if I recieve enough feedback for features via the appropriate channels (such as my [HUD / Project Discord](https://discord.gg/HyZRVtp)) I will do my best to add them.
| Key Word       |                          Status                           | Description                                                                        |
|----------------|:---------------------------------------------------------:|------------------------------------------------------------------------------------|
| CreateShortcut |                          Planned                          | Create a shortcut to launch DemoArk with the current options, minus CreateShortcut |
| LaunchTF2      |                          Planned                          | Launch TF2 after running DemoArk                                                   |
| SortMonth      |  Removed, too niche<br>(may return with adequate demand)  | Group demos into folders by month                                                  |
| TwelveHourTime | Removed, too finicky<br>(may return with adequate demand) | **True:** 12hr<br>**False:** 24hr                                                  |

# Is This A Virus?
### Nope.
Your anti-virus will likely warn you about running executables made by unknown creators, which is its job. If you know how, I invite you to review my code, build it from source, and even provide feedback.  
If the project takes off I can look into getting the project officially approved my Microsoft so Windows Security won't show the scary warning. Though to maintain Microsoft recognition I'd have to do it for every update or pay $300-$700 annually for Microsoft to keep track of it automatically. This also wouldn't apply to third-party anti-virus providers.

# Thanks
**[NeKzor's Portal 2 Demo Documentation](https://dem.nekz.me/)** - The most comprehensive documentation on demos I've ever seen, even if for a different game.  
**[Klauspost's Compress Go Package](https://github.com/klauspost/compress)** - A drop-in 15% - 25% optimization to the Go Standard Library Zip package.  
**[Pektezol's BitReader Go Package](https://github.com/pektezol/bitreader)** - An accessible package for parsing bit-buffered demo contents in Go.  
**[Demostf's Demo Parser](https://codeberg.org/demostf/parser)** - The reference for TF2 specific bit-values, packet, and message specifications.  
**Ward** - Providing feedback and testing predecessor programs and pre-release versions of DemoArk.
