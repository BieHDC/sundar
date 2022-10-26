# Changelog
---

## Version 2
- You can now soft stop the bot with one ctrl+c and force stop it if you press ctrl+c a second time in case it hangs. Can happen when the homeserver goes down and such things.
- Removed the `cluster` system and instead provide a `roomlistcommand`. Usage is simply `roomlistcommand room1,room2,room3 thecommand args`. It was barely used and had little value for the amount of code it created.
- Added a function to replace youtube short links with a normal watch link, and optionally reply with a link to a piped instance. (its just a url replace)
- Added a function to replace twitter links with nitter links. (also just a url replace)
- Added a function and command to generate a search url. You can ether type `!ddg the search term` or you reply to an existing message with `!ddg` (no extra words at the end) to turn that message content into a search term.
- Added a lot more dad jokes and quotes.
- Echos are not runtime anymore.
In the subfolder `echos` there is now an `echos.csv` with the format `callsign,powerlevel,"the message"`. To insert a newline, use `??NL??`.
When you are done, run the `main.go` file like `go run main.go -in echos.csv -out echos.json` to generate a new echolist. You then pass this `echos.json` into the bot itself with the `-echos` launchflag. If none provided, it will not load any echos at all.
>
- The guess the number game is now per room instead of global.
- Reworked the random message implementation for dad jokes and quotes to be much better and extendable.
- Cleaned up the snarkyReplies functionality.
- The Windows Error Code files are now being parsed at compile time by `error_codes/generate_codes.go` and then included into the source code.
- Made a new system for commands and function to send messages in order to have them better testable. The direct functions can still be used.
- Some functions and commands are now closures, so they initialise themselves and then return a function, making initial setups much cleaner and removes a lot of global state.
- General Code overhaul here and there.
- Added tests for functions where i thought it makes sense. Same for fuzzing.


---
## Version 1
 - Initial testing version.