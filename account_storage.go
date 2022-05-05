package main

import (
    "maunium.net/go/mautrix"
	"errors"
	"fmt"
	"sync"
	"encoding/json"
)

type AccountStorage struct {
	storage_prefix 	string
	storage_mutex 	sync.Mutex
	storage_cache 	map[string][]byte
	storage_devmode bool // when in devmode we dont flush storage requests to the server
}

func NewDefaultAccountStorage(prefix string, devmode bool) *AccountStorage {
	if devmode == true {
		fmt.Println("!!STORAGE DEV MODE ACTIVE: NO DATA WILL BE SAVED ON THE SERVER!!")
	}
	return &AccountStorage{
		storage_prefix: prefix,
		storage_cache: 	make(map[string][]byte),
		storage_devmode: devmode,
	}
}

func (cmdhdlr *CommandHandler) StoreData(key string, data interface{}) {
	cmdhdlr.accountstore.storage_mutex.Lock()
	defer cmdhdlr.accountstore.storage_mutex.Unlock()

	if cmdhdlr.accountstore.storage_devmode == false {
		err := cmdhdlr.client.SetAccountData(cmdhdlr.accountstore.storage_prefix+key, &data)
		if err != nil { 
			BotNotifyEventsChannel(cmdhdlr, "StoreData >" + key + "< Error:" + err.Error())
		} else {
			cmdhdlr.accountstore.storage_cache[cmdhdlr.accountstore.storage_prefix+key], err = json.Marshal(data)
			if err != nil {
				BotNotifyEventsChannel(cmdhdlr, "\tFAILED to cache1 >" + key + "<:" + fmt.Sprintf("%+v", data))
			} //else {
			//	fmt.Printf("\tcaching %+v\n", data)
			//}
		}
	} else { //DEVMODE: CACHE ONLY
		fmt.Printf("Store for %s\n", key)
		var err error
		cmdhdlr.accountstore.storage_cache[cmdhdlr.accountstore.storage_prefix+key], err = json.Marshal(data)
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "\tFAILED to cache2 >" + key + "<:" + fmt.Sprintf("%+v", data))
		}
	}
}

func (cmdhdlr *CommandHandler) FetchData(key string, data interface{}) {
	cmdhdlr.accountstore.storage_mutex.Lock()
	defer cmdhdlr.accountstore.storage_mutex.Unlock()

	if cmdhdlr.accountstore.storage_devmode == true {
		fmt.Printf("Fetch for %s\n", key)
	}
	cached_data, exists := cmdhdlr.accountstore.storage_cache[cmdhdlr.accountstore.storage_prefix+key]
	if exists {
		err := json.Unmarshal(cached_data, data)
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "\tFAILED to get cache2 >" + key + "<:" + fmt.Sprintf("%+v", cached_data))
		}// else {
		//	fmt.Printf("\tfrom cache %+v\n", data)	
		//}
	} else {
		if cmdhdlr.accountstore.storage_devmode == false { //DEVMODE: Dont fetch since we never flush anyway
			err := cmdhdlr.client.GetAccountData(cmdhdlr.accountstore.storage_prefix+key, &data)
			if err != nil && !errors.Is(err, mautrix.MNotFound) {
				BotNotifyEventsChannel(cmdhdlr, "FetchData >" + key + "< Error:" + err.Error())
			} else {
				cmdhdlr.accountstore.storage_cache[cmdhdlr.accountstore.storage_prefix+key], err = json.Marshal(data)
				if err != nil {
					BotNotifyEventsChannel(cmdhdlr, "\tFAILED to cache3 >" + key + "<:" + fmt.Sprintf("%+v", data))
				}
				//fmt.Printf("\tfrom cold %+v\n", data)
			}
		} else {
			fmt.Println("!!STORAGE DEV MODE: TRYING TO FETCH NON EXISTING KEY >", key, "<. Must not be an error!!")
		}
	}
}
