# Safed Sundarata - A Matrix Bot written in Go
Safed Sundarata, or in short Sundar, is a Matrix bot with a big variety of functionality. The name was chosen by the first person who answered when i asked for one, because i am not good with naming (see code). It is currently being used by the [ReactOS](https://reactos.org/) Community [on Matrix](https://matrix.to/#/#community:reactos.org) with it's main purpose of providing the `!error <code>` command for developers to look up Windows Error Messages.


---
## Usage
Copy `config.yaml.example` and edit it's content with your data and preferences.
```go
go build
./biehdc.safed_sundarata -config config.yaml
 ```
Invite the bot, like you would invite any other user, into the rooms it should exist in.
It will then introduce itself and i suggest you run the `help` command to find out what it can do.

---
## Features
### Commands
#### Custom Commands
- `error` convers a Windows Error Message Code to a human readable string.
- `dj` tells a dadjoke.
- `quote` tells a quote.
- `god` is terry davis god functionality where you can ask for advice.
- `hi` greet the bot.
- `guess` simple guess the number game.
- `status` shows stats about your server.

#### Builtin Commands
- `help` displays all commands you are *allowed to* run, meaning the output depends on your powerlevel.
- `usage` displays the usage of a command.

- `die` shuts the bot down.
- `reboot` you can guess that one.
- `botinfo` prints the same greeting message as it does when joining a room.

- `echoadd` adds a command that just repeats a message you give it. Useful for explanations that repeat often and etc.
- `echodel` removes it.

- `setdisplayname/setavatar` to globally set the bots appearance.
- `setbotroomappearance` to set the appearance in a room only.

- `poweroverride[reset[all]]` to override the powerlevel required for a command per room and reset them.
- `blockcommand` is just a shortcut for `poweroverride` with a powerlevel of 100.

- `forceleave/forcejoin` Commands to walk around Matrix.
- `leaveemptyrooms` to make the bot leave rooms it is alone in.
- `broadcast` broadcasts a message to all joined rooms.

- `cluster` and it's subcommands to run commands against a cluster of rooms.
- `say` is meant to be combined with `cluster cmd` to broadcast a message in a room cluster.


### Functions
A Function is a "Command" that runs on every message. You can have functions with a blacklist(run, unless blacklisted in the room) or whitelist(do not run, unless whitelisted in the room).
- `funcstate` allows you to enable or disable a function in a room.
- `funcwiperooms` deletes all overrides for a function.
- `funclist` lists all registered functions.

#### Custom Functions
- `snarkyReplies` does some sudo/doas trolling.
- `emojireactor` scans any message for the emojis it contains and adds them as reaction.
- `userspecifictrolling` adds reactions to a specific users message.


#### Cron
You can also add Cronjob functions that run periodically.
- `CronServerStatus` prints what the `status` command would output every day at midnight into your events channel.

---
## Improvements
This is my first project written in Go and i probably did a lot of C-things. Feel free to suggest changes to code and structure, however please do not suggest changes for the sake of change or show-off code. I value readability and structure.

---
## Contributing
I am going to assume you are a nice person, and if you are not, i am going to be sad.

---
## Deploy your own
Unlike other projects, this one is meant for you to just fork it and make the changes you need. If you add general functionality, keep them in the designated files (misc utils, management, etc...) and put your very custom code (like the error code searcher) in files prefixed with `custom_`. This layout may be subject to change, but that is the general idea.

---
## Donations
If you want to Donate to this project, please directly donate to [ReactOS](https://reactos.org/donate/) and/or the person hosting the bot for the ReactOS Community [Cernodile](https://cernodile.com/donate.php). This way i do not have to redirect the funds manually and please add a reference like "matrixbot" to your donation platform choice just to see how much of them came through these links.

---
## License
AGPL3 plus Additions listed in the LICENSE file.