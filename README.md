# Demo Ark
The automated demo archiving program for TF2 to help save you from the flood of demos recorded by ds_enable.

# Options
Demo Ark can be run with a slew of options, configured either via the prompts upon running the program, or via arguments used to launch the program through the command line or similar means.  
Arguments are should be separated by spaces, with each argument consisting of a keyword, equal sign, and value (no spaces). Capitalization is not important. Boolean values can be given as true/false or 1/0.  
The IgnoreWords argument is the exception to this formatting, it should be the last argument, with all desired ignored words being separated with spaces.  
For example:  
`./demo-ark silent=true sortYear=1 ZipOlderThan=3 SNIPE=the_med.dem IgnoreWords untouchable secret`  
| Key Word       |     Possible Values      | Default Value | Description                                                                  |
|----------------|:------------------------:|:-------------:|------------------------------------------------------------------------------|
| Silent         |   true (1) / false (0)   |     false     | Run the program without prompts                                              |
| SortYear       |   true (1) / false (0)   |     true      | Group demos into folders by year                                             |
| SortMonth      |   true (1) / false (0)   |     false     | Group demos into folders by month                                            |
| SortGameType   |   true (1) / false (0)   |     true      | Group demos into folders by gametype                                         |
| KeepPrefix     |   true (1) / false (0)   |     true      | Rename options won't overwrite a detected ds_prefix                          |
| RenameMap      |   true (1) / false (0)   |     false     | Rename the demo to contain the map name                                      |
| RenameDuration |   true (1) / false (0)   |     false     | Rename demo to contain the duration of the demo                              |
| SearchDirs     |   true (1) / false (0)   |     false     | Search subdirectories within the current directory                           |
| Multithread    |   true (1) / false (0)   |     true      | Allow the use of multiple cores / threads                                    |
| DateMajorDir   |   true (1) / false (0)   |     true      | True: year/month/gametype/demo.dem False: gametype/year/month/demo.dem       |
| UseEditDate    |   true (1) / false (0)   |     false     | Use the date that a demo was last edited rather than date in its file name   |
| TwelveHourTime |   true (1) / false (0)   |     false     | True: 12hr False: 24hr                                                       |
| SetAsideCulled |   true (1) / false (0)   |     true      | Set aside culled demos to a "culled" directory, rather than deleting them    |
| ShowConVars    |   true (1) / false (0)   |     false     | Outputs console variables parsed from demo (mainly for debugging)            |
| ZipOlderThan   |    0 (disabled) - 255    |       1       | Zip demos older than this many years                                         |
| CullBelow      |    0 (disabled) - 255    |      10       | Number of seconds that demos below that duration will be deleted             |
| Snipe          | Any continuous filename  |      ""       | Snipe a specific file (exactly) to execute program on (mainly for debugging) |
| IgnoreWords    | Any words after key word |  "reference"  | Ignore file / folder names containing string                                 |
