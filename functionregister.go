package main

import (
	//"fmt"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	//"strings"
	"sync"
)

const (
	StatusUnknown   int = 0
	StatusBlacklist int = 1
	StatusWhitelist int = 2
)

type FunctionRegisterPrototype func(*CommandHandler, id.RoomID, id.UserID, int, []string, *event.Event)

type FunctionRegister struct {
	allfunctions 			map[string]FunctionRegisterPrototype
	functionstatus 			map[string]int
	Functionrooms 			map[string][]id.RoomID
	func_register_mutex 	sync.Mutex
}

//AccountStorage will take care of caching
func saveFunctionstatus(cmdhdlr *CommandHandler, funcregister *FunctionRegister) {
    cmdhdlr.StoreData("funcregister_2", funcregister)
}

func loadFunctionstatus(cmdhdlr *CommandHandler, funcregister *FunctionRegister) {
    cmdhdlr.FetchData("funcregister_2", funcregister)
}

func NewDefaultFunctionRegister() *FunctionRegister {
	return &FunctionRegister{
		allfunctions: 	make(map[string]FunctionRegisterPrototype),
		functionstatus:	make(map[string]int),
		Functionrooms:	make(map[string][]id.RoomID),
	}
}

func (cmdhdlr* CommandHandler) AddFunction(functionid string, status int, targetfunction FunctionRegisterPrototype) {
	cmdhdlr.funcregister.allfunctions[functionid] = targetfunction
	cmdhdlr.funcregister.functionstatus[functionid] = status
}


func HandleFunctionWipeRoomlist(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.funcregister.func_register_mutex.Lock()
	defer cmdhdlr.funcregister.func_register_mutex.Unlock()
	
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		_, exists := cmdhdlr.funcregister.allfunctions[argv[1]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "The function >" + argv[1] + "< does not exist.")
			return false
		}

		newroomlist := []id.RoomID{}
		cmdhdlr.funcregister.Functionrooms[argv[1]] = newroomlist
		saveFunctionstatus(cmdhdlr, cmdhdlr.funcregister)

		BotReplyMsg(cmdhdlr, statusroom, "The roomlist for the function >" + argv[1] + "< has been deleted.")
		return true
	}
}

func HandleFunctionStatusInRoom(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.funcregister.func_register_mutex.Lock()
	defer cmdhdlr.funcregister.func_register_mutex.Unlock()
	
	if argc < 3 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		// argv 1 -> functionid
		// argv 2 -> enable/disable
		// argv 3 -> roomid
		var targetroom_str string
		if argv[3] == "this" {
			targetroom_str = room.String()
		} else {
			targetroom_str = argv[3]
		}

		targetroom, err := parseRoomID(targetroom_str)
		if err != nil || targetroom == nil {
    		BotReplyMsg(cmdhdlr, statusroom, "Invalid RoomID.")
			return false
		}

		if !(cmdhdlr.userHasPowerlevel(targetroom.RoomID(), sender) >= CommandAdmin) {
			BotReplyMsg(cmdhdlr, statusroom, "You need to be an Administrator in >" + targetroom.String() + "< to block functions.")
			return true
		}

		newstate := false
		if argv[2] == "enable" {
			newstate = true
		} else if argv[2] == "disable" {
			newstate = false
		} else {
    		BotReplyMsg(cmdhdlr, statusroom, "New State must be <enable> or <disable>.")
			return false
		}

		_, exists := cmdhdlr.funcregister.allfunctions[argv[1]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "The function >" + argv[1] + "< does not exist.")
			return false
		}

		functionstatus := cmdhdlr.funcregister.functionstatus[argv[1]]
		finalstatus := "Undefined, should not happen. Report to dev if it does!"
		// blacklist & enable -> remove from room list
		// whitelist & disable -> remove from room list
		if ((functionstatus == StatusBlacklist) && (newstate == true)) || ((functionstatus == StatusWhitelist) && (newstate == false)) {
			var newroomlist []id.RoomID
			// Check if the room is in the list
			roomfound := false
			functionrooms, exists := cmdhdlr.funcregister.Functionrooms[argv[1]]
			if exists {
				for _, roomlistentry := range functionrooms {
					if roomlistentry == targetroom.RoomID() {
						roomfound = true
					} else {
						// Append rooms that are not the target
						newroomlist = append(newroomlist, roomlistentry)
					}
				}
			}

			if roomfound {
				oldfunctionrooms := cmdhdlr.funcregister.Functionrooms[argv[1]]
				oldfunctionrooms = newroomlist
				cmdhdlr.funcregister.Functionrooms[argv[1]] = oldfunctionrooms

				if functionstatus == StatusBlacklist {
					finalstatus = "Function >" + argv[1] + "< has been unblocked in >" + targetroom.String() + "<."
				} else {
					finalstatus = "Function >" + argv[1] + "< has been blocked in >" + targetroom.String() + "<."
				}
			} else {
				//we didnt have to do anything anyway
				if functionstatus == StatusBlacklist {
					finalstatus = "Function >" + argv[1] + "< was not blocked in >" + targetroom.String() + "<."
				} else {
					finalstatus = "Function >" + argv[1] + "< was already blocked in >" + targetroom.String() + "<."
				}
				BotReplyMsg(cmdhdlr, statusroom, finalstatus)
				return true
			}

		// blacklist & disable -> add to room list
		// whitelist & enable -> add to room list
		} else if ((functionstatus == StatusBlacklist) && (newstate == false)) || ((functionstatus == StatusWhitelist) && (newstate == true)) {
			// Check if its not already in the list
			functionrooms, exists := cmdhdlr.funcregister.Functionrooms[argv[1]]
			if exists {
				for _, roomlistentry := range functionrooms {
					if roomlistentry == targetroom.RoomID() {
						// if its already in the list, we are done and can exit cleanly
						if functionstatus == StatusBlacklist {
							finalstatus = "Function >" + argv[1] + "< was already blocked in >" + targetroom.String() + "<."
						} else {
							finalstatus = "Function >" + argv[1] + "< was already unblocked in >" + targetroom.String() + "<."
						}
						BotReplyMsg(cmdhdlr, statusroom, finalstatus)
						return true
					}
				}
			}

			// Otherwise add it
			oldfunctionrooms := cmdhdlr.funcregister.Functionrooms[argv[1]]
			oldfunctionrooms = append(oldfunctionrooms, targetroom.RoomID())
			cmdhdlr.funcregister.Functionrooms[argv[1]] = oldfunctionrooms
			if functionstatus == StatusBlacklist {
				finalstatus = "Function >" + argv[1] + "< has been blocked in >" + targetroom.String() + "<."
			} else {
				finalstatus = "Function >" + argv[1] + "< has been unblocked in >" + targetroom.String() + "<."
			}
		}

		// If we got this far, we have to save
		saveFunctionstatus(cmdhdlr, cmdhdlr.funcregister)

		BotReplyMsg(cmdhdlr, statusroom, finalstatus)
	}
	return true
}


func HandleListKnownFunctions(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	list := "Registered Functions: "
    if len(cmdhdlr.funcregister.allfunctions) > 0 {
    	list += "\n"
	    for fregname, _ := range cmdhdlr.funcregister.allfunctions {
	    	list += "\t" + fregname
	    	if cmdhdlr.funcregister.functionstatus[fregname] == StatusBlacklist {
	    		list += " -> Blacklist: "
	    	} else {
	    		list += " -> Whitelist: "
	    	}
	    	if len(cmdhdlr.funcregister.Functionrooms[fregname]) > 0 {
		    	for _, targetroom := range cmdhdlr.funcregister.Functionrooms[fregname] {
		    		list += targetroom.String() + " "
		    	}
	    	} else {
	    		list += "None"
	    	}
	    	list += "\n"
	    }
	} else {
		list += "None"
	}
    BotReplyMsg(cmdhdlr, statusroom, list)
	return true
}


func (freg *FunctionRegister) InvokeFunctions(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) {
	for fregname, fregfun := range freg.allfunctions {
		has_entry := false
		functionrooms, exists := freg.Functionrooms[fregname]
		if exists {
			for _, roomlistentry := range functionrooms {
				if roomlistentry == room {
					has_entry = true
					break
				}
			}
		}
		// If the function is whitelist based and has an entry OR the function is blacklist based and doesnt have an entry -> exec
		if ((freg.functionstatus[fregname] == StatusWhitelist) && has_entry) || ((freg.functionstatus[fregname] == StatusBlacklist) && !has_entry) {
			fregfun(cmdhdlr, room, sender, argc, argv, evt)
		}
	}
}

