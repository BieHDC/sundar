package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	cron "github.com/robfig/cron/v3"

	//"github.com/pkg/profile" // For Profiling
)

type BotConfiguration struct {
	// Authentication settings
	Homeserver              string `yaml:"homeserver"`
	LoginType               string `yaml:"login_type"` // "password" or "accesstoken"
	Username_or_Userid      string `yaml:"username_or_userid"`
	Password_or_Accesstoken string `yaml:"password_or_accesstoken"`

	// Bot settings
	Prefix            	string    	`yaml:"prefix"`
	DeviceDisplayName 	string    	`yaml:"device_display_name"`
	UniqueDeviceID    	string    	`yaml:"unique_device_id"`
	Master            	id.UserID 	`yaml:"master"`
	EventsChannel     	id.RoomID 	`yaml:"events_channel"`
	LimitInvites      	bool      	`yaml:"limit_invites"`
	AccountStorePrefix 	string 		`yaml:"accountstore_prefix"`
	AccountStoreDebug 	bool 		`yaml:"accountstore_debug"`
}

func (config *BotConfiguration) Parse(data []byte) error {
	return yaml.UnmarshalStrict(data, config)
}

// Make sure we can exit cleanly
var closechannel = make(chan os.Signal, 1)
var bot_is_quitting = false
var bot_is_rebooting = false

func main() {
	//defer profile.Start().Stop() // For Profiling
	configPath := flag.String("config", "", "config.yaml file location")
	flag.Parse()
	if *configPath == "" {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	fmt.Printf("Reading config from %s...\n", *configPath)
	configYaml, err := os.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	botconfig := BotConfiguration{}
	err = botconfig.Parse(configYaml)
	if err != nil {
		panic(err)
	}

	if botconfig.Homeserver == "" {
		panic("No Homeserver specified in your config file")
	}
	if botconfig.Username_or_Userid == "" || botconfig.Password_or_Accesstoken == "" {
		panic("Empty login data")
	}

	var client *mautrix.Client
	if botconfig.LoginType == "password" {
		fmt.Println("Logging into", botconfig.Homeserver, "as", botconfig.Username_or_Userid)
		client, err = mautrix.NewClient(botconfig.Homeserver, "", "")
		if err != nil {
			panic(err)
		}

		serverreply, err := client.Login(&mautrix.ReqLogin{
			Type:                     mautrix.AuthTypePassword,
			Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: botconfig.Username_or_Userid},
			Password:                 botconfig.Password_or_Accesstoken,
			StoreCredentials:         true,
			DeviceID:                 id.DeviceID(botconfig.UniqueDeviceID),
			InitialDeviceDisplayName: botconfig.DeviceDisplayName,
		})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Consider using Access token based login:\n\tUserID:%s\n\tAccessToken:%s\n", serverreply.UserID, serverreply.AccessToken)
		fmt.Println("Login successful")
	} else if botconfig.LoginType == "accesstoken" {
		client, err = mautrix.NewClient(botconfig.Homeserver, id.UserID(botconfig.Username_or_Userid), botconfig.Password_or_Accesstoken)
		if err != nil {
			panic(err)
		}
		client.DeviceID = id.DeviceID(botconfig.UniqueDeviceID)
		fmt.Printf("New client created\n")
	} else {
		panic("Unknown login type " + botconfig.LoginType)
	}

	// Technically when using Access Token login we could use the config value, but this is a nice initial test if the connection is actually successful
	resp, err := client.Whoami()
	if err != nil {
		panic(err)
	}
	our_uid := resp.UserID.String()

	cronhanlder := cron.New()
	botinfofunc := func(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) (bool) {
			BotReplyMsg(cmdhdlr, evt.RoomID, 
				"Hello, i am " + botconfig.DeviceDisplayName +
				". \nYou can address me with >" + botconfig.Prefix +
				"< and if you type >" + botconfig.Prefix +
				"help< i can show you what i can do.\n" +
				"If there are any issues with me, please tell my master "+botconfig.Master.String() + ".\n" + 
				"You can download my AGPL3 Source Code at https://github.com/BieHDC/sundar")
			return true
	}


	cmdhdlr := NewCommandHandler(client, botconfig.Prefix, botconfig.Master, 
									botconfig.EventsChannel, true, 
									botconfig.AccountStorePrefix, botconfig.AccountStoreDebug)

	// Your custom Commands
	// Productivity
	cmdhdlr.AddCommand("error", "Get Friendly Name of Windows Error", "<Error Code>", "Productivity", CommandAnyone, HandleErrorCodeRequest)
	// Fun
	cmdhdlr.AddCommand("dj", "Tell random Dadjoke", 	"", 			"Fun", CommandAnyone, HandleRandomDadjoke)
	cmdhdlr.AddCommand("quote", "Tell random Quote", 	"", 			"Fun", CommandAnyone, HandleRandomQuote)
	cmdhdlr.AddCommand("god", "Ask God for advice", 	"<your words>", "Fun", CommandAnyone, HandleGod)
	cmdhdlr.AddCommand("hi", "Say Hi to the bot", 		"", 			"Misc", CommandAnyone, HandleHello)
	cmdhdlr.AddCommand("guess", "Guess the number", 	"<number>", 	"Games", CommandAnyone, HandleGuess)
	// Serverinfo
	cmdhdlr.AddCommand("status", "Print the current Server Status", "", "Misc", CommandModerator, HandleServerStatus)

	// Your custom Functions
	cmdhdlr.AddFunction("snarkyReplies", StatusBlacklist, snarkyReplies)
	cmdhdlr.AddFunction("emojireactor", StatusWhitelist, emojireactor)
	cmdhdlr.AddFunction("goodoldfriend", StatusBlacklist, goodoldfriend)

	// Your CRON Jobs
	cronhanlder.AddFunc("00 00 * * *", func() { CronServerStatus(cmdhdlr) })

	// These 2 go here because those events are also handled here
	cmdhdlr.AddCommand("die", "Shut down the bot", "", "Bot Administration", CommandMaster, HandleCleanShutdown)
	cmdhdlr.AddCommand("reboot", "Reboot the bot", "", "Bot Administration", CommandMaster, HandleReboot)
	cmdhdlr.AddCommand("botinfo", "Display Botinfo", "", "Help", CommandAnyone, botinfofunc)
	// This should be called after everything is done
	cmdhdlr.Done()
	cronhanlder.Start()


	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	ignoreoldevents := mautrix.OldEventIgnorer{UserID: client.UserID}
	ignoreoldevents.Register(syncer)

	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		// Ignore messages from ourselves
		if evt.Sender == client.UserID {
			return
		}
		// This prints every message to your console, if you need to check around
		// fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)(room>>%[5]s<<)(%[6]d)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body, evt.RoomID, evt.Timestamp)

		go cmdhdlr.HandleDispatch(evt)
	})

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if *evt.StateKey == our_uid && evt.Content.AsMember().Membership == event.MembershipInvite {
			time.Sleep(1 * time.Second)
			BotNotifyEventsChannel(cmdhdlr, "Join Event:"+evt.RoomID.String())

			if botconfig.LimitInvites == true && evt.Sender != botconfig.Master {
				BotNotifyEventsChannel(cmdhdlr, "User >"+evt.Sender.String()+"< tried to invite the bot to room >"+
					evt.RoomID.String()+"< but only the botmaster is allowed to invite!")
			} else {
				joinsuccess := botJoinRoom(cmdhdlr, evt.RoomID)
				if joinsuccess {
					botinfofunc(cmdhdlr, evt.RoomID, id.UserID(0), 0, []string{}, evt.RoomID, evt)
				}
			}
		} else if *evt.StateKey == our_uid && evt.Content.AsMember().Membership.IsLeaveOrBan() {
			cmdhdlr.leave_room_event(evt.RoomID)
			BotNotifyEventsChannel(cmdhdlr, "Left or banned from:"+evt.RoomID.String())
		}
	})

	// Watch for power level changes as it affects who can run which bot commands
	syncer.OnEventType(event.StatePowerLevels, func(source mautrix.EventSource, evt *event.Event) {
		cmdhdlr.PowerLevelsChange(evt.RoomID, evt.Content.AsPowerLevels().Users)
	})

	// Debug Function to see whats stored in account data
/*
		syncer.OnSync(func(resp *mautrix.RespSync, since string) bool{
			if len(resp.AccountData.Events) > 0 {
				fmt.Printf("acc: %+v\n", resp.AccountData.Events[0].Content.Raw)
			}
			return true
		})
*/	

	// Make sure we can exit cleanly
	signal.Notify(closechannel,
		os.Interrupt,
		os.Kill,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	go func() {
		for range closechannel {
			if bot_is_rebooting {
				BotNotifyEventsChannel(cmdhdlr, "Bot is rebooting")
			} else {
				BotNotifyEventsChannel(cmdhdlr, "Bot is shutting down")
			}
			bot_is_quitting = true
			client.StopSync()
			time.Sleep(1 * time.Second)
			client.Client.CloseIdleConnections()

			if bot_is_rebooting {
				argv0, err := exec.LookPath(os.Args[0])
				if err != nil {
					fmt.Println("Error while rebooting the bot, shutting down instead: ", err.Error())
					break
				}
				_, err = os.Stat(argv0)
				if err != nil {
					fmt.Println("Error while rebooting the bot, shutting down instead: ", err.Error())
					break
				}

				syscall.Exec(argv0, os.Args, os.Environ())
			}
		}
		os.Exit(0)
	}()

	BotNotifyEventsChannel(cmdhdlr, "Bot started, running sync loop...")
	for {
		err = client.Sync()
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "Sync Error:"+err.Error())
		}
		if bot_is_quitting {
			break
		}
	}
	// We need to clog up the pipe or we wont make it to syscall.Exec
	if bot_is_rebooting {
		for {}
	}
}

func HandleCleanShutdown(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	closechannel <- os.Interrupt
	return true
}

func HandleReboot(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	bot_is_rebooting = true
	return HandleCleanShutdown(cmdhdlr, room, sender, argc, argv, statusroom, evt)
}