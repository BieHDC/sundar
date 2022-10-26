package main

import (
	//"fmt"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	//"strings"
	"sync"
)

type functionStatus int
const (
	StatusUnknown functionStatus = iota
	StatusBlacklist
	StatusWhitelist
)

type FunctionRegisterPrototype func(*CommandHandler, id.RoomID, id.UserID, int, []string, *event.Event) BotReply

type FunctionRegister struct {
	allfunctions 			map[string]FunctionRegisterPrototype
	functionstatus 			map[string]functionStatus
	Functionrooms 			map[string][]id.RoomID
	func_register_mutex 	sync.Mutex
}

//AccountStorage will take care of caching
func saveFunctionstatus(cmdhdlr *CommandHandler, funcregister *FunctionRegister) {
    cmdhdlr.accountstore.StoreData(cmdhdlr.client, "funcregister_2", funcregister)
}

func loadFunctionstatus(cmdhdlr *CommandHandler, funcregister *FunctionRegister) {
    cmdhdlr.accountstore.FetchData(cmdhdlr.client, "funcregister_2", funcregister)
}

func NewDefaultFunctionRegister() *FunctionRegister {
	return &FunctionRegister{
		allfunctions: 	make(map[string]FunctionRegisterPrototype),
		functionstatus:	make(map[string]functionStatus),
		Functionrooms:	make(map[string][]id.RoomID),
	}
}


func (freg *FunctionRegister) InvokeFunctions(ca CommandArgs) {
	for fregname, fregfun := range freg.allfunctions {
		has_entry := false
		functionrooms, exists := freg.Functionrooms[fregname]
		if exists {
			for _, roomlistentry := range functionrooms {
				if roomlistentry == ca.room {
					has_entry = true
					break
				}
			}
		}

		// If the function is whitelist based and has an entry OR the function is blacklist based and doesnt have an entry -> exec
		if ((freg.functionstatus[fregname] == StatusWhitelist) && has_entry) || ((freg.functionstatus[fregname] == StatusBlacklist) && !has_entry) {
			ca.cmdhdlr.BotPrint(fregfun(ca.cmdhdlr, ca.room, ca.sender, ca.argc, ca.argv, ca.evt))
		}
	}
}


func (cmdhdlr* CommandHandler) AddFunction(functionid string, status functionStatus, targetfunction FunctionRegisterPrototype) {
	cmdhdlr.funcregister.allfunctions[functionid] = targetfunction
	cmdhdlr.funcregister.functionstatus[functionid] = status
}


func HandleFunctionWipeRoomlist(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	ca.cmdhdlr.funcregister.func_register_mutex.Lock()
	defer ca.cmdhdlr.funcregister.func_register_mutex.Unlock()
	
	_, exists := ca.cmdhdlr.funcregister.allfunctions[ca.argv[1]]
	if !exists {
		return BotPrintSimple(ca.statusroom, "The function >" + ca.argv[1] + "< does not exist.")
	}

	newroomlist := []id.RoomID{}
	ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]] = newroomlist
	saveFunctionstatus(ca.cmdhdlr, ca.cmdhdlr.funcregister)

	return BotPrintSimple(ca.statusroom, "The roomlist for the function >" + ca.argv[1] + "< has been deleted.")
}

func HandleFunctionStatusInRoom(ca CommandArgs) BotReply {
	if ca.argc < 3 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}
	
	// argv 1 -> functionid
	// argv 2 -> enable/disable
	// argv 3 -> roomid
	var targetroom_str string
	if ca.argv[3] == "this" {
		targetroom_str = ca.room.String()
	} else {
		targetroom_str = ca.argv[3]
	}

	targetroom, err := parseRoomID(targetroom_str)
	if err != nil || targetroom == nil {
		return BotPrintSimple(ca.statusroom, "Invalid RoomID.")
	}

	if !(ca.cmdhdlr.userHasPowerlevel(targetroom.RoomID(), ca.sender) >= CommandAdmin) {
		return BotPrintSimple(ca.statusroom, "You need to be an Administrator in >" + targetroom.String() + "< to block functions.")
	}

	newstate := false
	if ca.argv[2] == "enable" {
		newstate = true
	} else if ca.argv[2] == "disable" {
		newstate = false
	} else {
		return BotPrintSimple(ca.statusroom, "New State must be <enable> or <disable>.")
	}

	ca.cmdhdlr.funcregister.func_register_mutex.Lock()
	defer ca.cmdhdlr.funcregister.func_register_mutex.Unlock()

	_, exists := ca.cmdhdlr.funcregister.allfunctions[ca.argv[1]]
	if !exists {
		return BotPrintSimple(ca.statusroom, "The function >" + ca.argv[1] + "< does not exist.")
	}

	functionstatus := ca.cmdhdlr.funcregister.functionstatus[ca.argv[1]]
	finalstatus := "Undefined, should not happen. Report to dev if it does!"
	// blacklist & enable -> remove from room list
	// whitelist & disable -> remove from room list
	if ((functionstatus == StatusBlacklist) && (newstate == true)) || ((functionstatus == StatusWhitelist) && (newstate == false)) {
		var newroomlist []id.RoomID
		// Check if the room is in the list
		roomfound := false
		functionrooms, exists := ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]]
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
			ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]] = newroomlist

			if functionstatus == StatusBlacklist {
				finalstatus = "Function >" + ca.argv[1] + "< has been unblocked in >" + targetroom.String() + "<."
			} else {
				finalstatus = "Function >" + ca.argv[1] + "< has been blocked in >" + targetroom.String() + "<."
			}
			saveFunctionstatus(ca.cmdhdlr, ca.cmdhdlr.funcregister)
			return BotPrintSimple(ca.statusroom, finalstatus)
		} else {
			//we didnt have to do anything anyway
			if functionstatus == StatusBlacklist {
				finalstatus = "Function >" + ca.argv[1] + "< was not blocked in >" + targetroom.String() + "<."
			} else {
				finalstatus = "Function >" + ca.argv[1] + "< was already blocked in >" + targetroom.String() + "<."
			}
			return BotPrintSimple(ca.statusroom, finalstatus)
		}

	// blacklist & disable -> add to room list
	// whitelist & enable -> add to room list
	} else if ((functionstatus == StatusBlacklist) && (newstate == false)) || ((functionstatus == StatusWhitelist) && (newstate == true)) {
		// Check if its not already in the list
		functionrooms, exists := ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]]
		if exists {
			for _, roomlistentry := range functionrooms {
				if roomlistentry == targetroom.RoomID() {
					// if its already in the list, we are done and can exit cleanly
					if functionstatus == StatusBlacklist {
						finalstatus = "Function >" + ca.argv[1] + "< was already blocked in >" + targetroom.String() + "<."
					} else {
						finalstatus = "Function >" + ca.argv[1] + "< was already unblocked in >" + targetroom.String() + "<."
					}
					return BotPrintSimple(ca.statusroom, finalstatus)
				}
			}
		}

		// Otherwise add it
		oldfunctionrooms := ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]]
		oldfunctionrooms = append(oldfunctionrooms, targetroom.RoomID())
		ca.cmdhdlr.funcregister.Functionrooms[ca.argv[1]] = oldfunctionrooms
		if functionstatus == StatusBlacklist {
			finalstatus = "Function >" + ca.argv[1] + "< has been blocked in >" + targetroom.String() + "<."
		} else {
			finalstatus = "Function >" + ca.argv[1] + "< has been unblocked in >" + targetroom.String() + "<."
		}
		saveFunctionstatus(ca.cmdhdlr, ca.cmdhdlr.funcregister)
		return BotPrintSimple(ca.statusroom, finalstatus)

	// I cant think if a case this might happen, but that doesnt mean it couldnt
	} else {
		return BotPrintSimple(ca.statusroom, "SOMETHING WENT VERY WRONG IN THE FUNCTION REGISTER, REPORT TO DEV FOR INSPECTION!")
	}
}


func HandleListKnownFunctions(ca CommandArgs) BotReply {
	list := "Registered Functions: "
    if len(ca.cmdhdlr.funcregister.allfunctions) > 0 {
    	list += "\n"
	    for fregname, _ := range ca.cmdhdlr.funcregister.allfunctions {
	    	list += "\t" + fregname
	    	if ca.cmdhdlr.funcregister.functionstatus[fregname] == StatusBlacklist {
	    		list += " -> Blacklist: "
	    	} else {
	    		list += " -> Whitelist: "
	    	}
	    	if len(ca.cmdhdlr.funcregister.Functionrooms[fregname]) > 0 {
		    	for _, targetroom := range ca.cmdhdlr.funcregister.Functionrooms[fregname] {
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

    return BotPrintSimple(ca.statusroom, list)
}
