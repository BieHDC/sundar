package main

import (
    //"maunium.net/go/mautrix"
    event "maunium.net/go/mautrix/event"
    id "maunium.net/go/mautrix/id"
    "strings"
    //"fmt"
    "sync"
)

type ClusterInternal struct {
    Rooms []id.RoomID 		//rooms in this cluster
    Admin id.UserID 		//people who can make changes to the cluster
}

type ClusterRegister struct {
	Allclusters 	map[string]ClusterInternal //Key: Clustername; Value Clusterstruct
	cluster_mutex 	sync.Mutex
}

func NewClusterHandler() *ClusterRegister {
    return &ClusterRegister{
    	Allclusters: make(map[string]ClusterInternal),
    }
}

//AccountStorage will take care of caching
func saveClusters(cmdhdlr *CommandHandler, clreg *ClusterRegister) {
    cmdhdlr.StoreData("clusterregister_1", clreg)
}

func loadClusters(cmdhdlr *CommandHandler, clreg *ClusterRegister) {
    cmdhdlr.FetchData("clusterregister_1", clreg)
}


func HandleCluster(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.clusterregister.cluster_mutex.Lock()
	defer cmdhdlr.clusterregister.cluster_mutex.Unlock()
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], evt.RoomID)
		return false
	} else {
		//fmt.Println(argv)
		//fmt.Printf("bef:")
		//fmt.Println(cmdhdlr.clusterregister.Allclusters)
		needssave := false
		switch argv[1] {
	        // available for all users
	        case "list":
	            needssave = handleClusterList(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        case "listroom":
	            needssave = handleClusterListroom(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        // needs clustermasters check
	        case "join":
	            needssave = handleClusterJoin(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        case "leave":
	            needssave = handleClusterLeave(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        case "remove":
	            needssave = handleClusterRemove(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        case "cmd":
	            needssave = handleClusterCmd(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, statusroom, evt)

	        default:
	            BotReplyMsg(cmdhdlr, statusroom, "Cluster command not found.")
	    }
		//fmt.Printf("aft:")
		//fmt.Println(cmdhdlr.clusterregister.Allclusters)

		if needssave {
			saveClusters(cmdhdlr, cmdhdlr.clusterregister)
		}
	}
	return true
}

func handleClusterList(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    list := "Known Clusters:"
    if len(cmdhdlr.clusterregister.Allclusters) > 0 {
    	list += "\n"
	    for name, cluster := range cmdhdlr.clusterregister.Allclusters {
	    	list += "\t" + name + " -> "
	    	for _, clusterroom := range cluster.Rooms {
	    		list += clusterroom.String() + " "
	    	}
	    	list += "\n"
	    }
	} else {
		list += " None"
	}
    BotReplyMsg(cmdhdlr, statusroom, list)
    return false
}

func handleClusterListroom(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    list := "This Room is member of the following Clusters: "
    if len(cmdhdlr.clusterregister.Allclusters) > 0 {
    	for name, cluster := range cmdhdlr.clusterregister.Allclusters {
	    	for _, clusterroom := range cluster.Rooms {
	    		if clusterroom == room {
	    			list += name + ", "
	    			break //it cant be twice in one cluster
	    		}
	    	}
	    }
	}

	if !strings.HasSuffix(list, ", ") {
		BotReplyMsg(cmdhdlr, statusroom, "This Room is not a member of any cluster.")
	} else {
    	BotReplyMsg(cmdhdlr, statusroom, strings.TrimSuffix(list, ", "))
	}

    return false
}

func handleClusterJoin(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    if argc < 4 {
		BotReplyMsg(cmdhdlr, statusroom, argv[0]+" "+argv[1]+" <clustername> <this>OR<roomid>")
	} else {
		// argv 2 -> clustername
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

		cluster, exists := cmdhdlr.clusterregister.Allclusters[argv[2]]
		if !exists {
			if cmdhdlr.userHasPowerlevel(targetroom.RoomID(), sender) >= CommandAdmin {
				cmdhdlr.clusterregister.Allclusters[argv[2]] = ClusterInternal{
					Rooms: []id.RoomID{targetroom.RoomID()},
					Admin: sender,
				}
    			BotReplyMsg(cmdhdlr, statusroom, "Created new cluster called >" + argv[2] + "<.")
   	 			return true
   	 		} else {
   	 			BotReplyMsg(cmdhdlr, statusroom, "You have to be an Administrator in the room you want to add to the cluster")
   	 			return false
   	 		}
		} else {
			//check if we are master in the cluster - fail
			if cluster.Admin == sender || sender == cmdhdlr.botmaster {
				//check if the room is already in the cluster - fail
				for _, clusterroom := range cluster.Rooms {
					if clusterroom == targetroom.RoomID() {
						BotReplyMsg(cmdhdlr, statusroom, "This room is already part of cluster >" + argv[2] + "<.")
						return false
					}
				}
				//add room to the cluster
				cluster.Rooms = append(cluster.Rooms, targetroom.RoomID())
				cmdhdlr.clusterregister.Allclusters[argv[2]] = cluster
				BotReplyMsg(cmdhdlr, statusroom, "Room has been added to the cluster >" + argv[2] + "<.")
				return true
			} else {
				BotReplyMsg(cmdhdlr, statusroom, "You are not a master in this cluster")
   				return false
			}
		}
	}
    return false
}

func handleClusterLeave(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    if argc < 4 {
		BotReplyMsg(cmdhdlr, statusroom, argv[0]+" "+argv[1]+" <clustername> <this>OR<roomid>")
	} else {
		// argv 2 -> clustername
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
		}

		cluster, exists := cmdhdlr.clusterregister.Allclusters[argv[2]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "Cluster called >" + argv[2] + "< does not exist.")
    		return false
		} 

		if cluster.Admin != sender || sender != cmdhdlr.botmaster {
			BotReplyMsg(cmdhdlr, statusroom, "You are not a master in this cluster")
    		return false
		}

		var clusterrooms []id.RoomID
		//check if the room is in the cluster
		roomfound := false
		for _, clusterroom := range cluster.Rooms {
			if clusterroom == targetroom.RoomID() {
				roomfound = true
			} else {
				// Append rooms that are not the target
				clusterrooms = append(clusterrooms, clusterroom)
			}
		}
		if roomfound {
			cluster.Rooms = clusterrooms
			cmdhdlr.clusterregister.Allclusters[argv[2]] = cluster
			BotReplyMsg(cmdhdlr, statusroom, "Room has been removed from the cluster >" + argv[2] + "<.")
			// If was last room in cluster, delete it
			if len(cluster.Rooms) < 1 {
				delete(cmdhdlr.clusterregister.Allclusters, argv[2])
				BotReplyMsg(cmdhdlr, statusroom, "Room was also the last in the cluster, removed it.")
			}
			return true
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "Room is not member of the cluster >" + argv[2] + "<.")
    		return false
		}
	}
    return false
}

func handleClusterRemove(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    if argc < 3 {
		BotReplyMsg(cmdhdlr, statusroom, argv[0]+" "+argv[1]+" <clustername>")
	} else {
		// argv 2 -> clustername
		cluster, exists := cmdhdlr.clusterregister.Allclusters[argv[2]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "Cluster called >" + argv[2] + "< does not exist.")
    		return false
		}

		if cluster.Admin == sender || sender == cmdhdlr.botmaster {
			delete(cmdhdlr.clusterregister.Allclusters, argv[2])
			BotReplyMsg(cmdhdlr, statusroom, "Removed cluster called >" + argv[2] + "<.")
			return true
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "You are not a master in this cluster")
    		return false
		}
	}
    return false
}

func handleClusterCmd(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
    if argc < 4 {
		BotReplyMsg(cmdhdlr, statusroom, argv[0]+" "+argv[1]+" <clustername> <command> [args]")
	} else {
		// argv 2  -> clustername
		// argv 3  -> command
		// argv 4+ -> command args
		cluster, exists := cmdhdlr.clusterregister.Allclusters[argv[2]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "Cluster called >" + argv[2] + "< does not exist.")
    		return false
		} 

		if cluster.Admin != sender || sender != cmdhdlr.botmaster {
			BotReplyMsg(cmdhdlr, statusroom, "You are not a master in this cluster")
    		return false
		}

		if argv[3] == argv[0] {
			BotReplyMsg(cmdhdlr, statusroom, "I will not let you recursively call this.")
    		return false
		} 

		//awesome magic, remove "cluster cmd" from argv
		newargv := argv[3:]
		newargc := len(newargv)
		for _, targetroom := range cluster.Rooms {
			success := InvokeCommand(cmdhdlr, targetroom, sender, newargc, newargv, statusroom, evt)
			if !success {
				BotReplyMsg(cmdhdlr, statusroom, "An error occurred while executing the command, stopping.")
				return false
			}
		}
		BotReplyMsg(cmdhdlr, statusroom, "Command executed in cluster >" + argv[2] + "<.")
    	return true
	}
    return false
}