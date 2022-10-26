package main

import (
    "maunium.net/go/mautrix"
	"errors"
	"fmt"
	"sync"
	"encoding/json"
)

type AccountStorage struct {
	storage_prefix 			string
	storage_mutex 			sync.Mutex
	storage_cache 			map[string][]byte
	storage_commitupstream 	bool // when in commitupstream we dont flush storage requests to the server
}

func NewDefaultAccountStorage(prefix string, commitupstream bool) *AccountStorage {
	// Note: the bool is inverted because the question in the config is different.
	if !commitupstream == false {
		fmt.Println("!!STORAGE DEV MODE ACTIVE: NO DATA WILL BE SAVED IN THE ACCOUNT DATA!!")
	}
	return &AccountStorage{
		storage_prefix: 		prefix,
		storage_cache: 			make(map[string][]byte),
		storage_commitupstream: !commitupstream,
	}
}


func (as *AccountStorage) StoreData(client *mautrix.Client, key string, data interface{}) error {
	as.storage_mutex.Lock()
	defer as.storage_mutex.Unlock()
	var err error

	as.storage_cache[as.storage_prefix+key], err = json.Marshal(data)
	if err != nil {
		err = errors.New("failed to cache for store for >" + key + "< with data >" + fmt.Sprintf("%+v", data) + "< and error: " + err.Error())
	}

	if as.storage_commitupstream == true {
		err = client.SetAccountData(as.storage_prefix+key, &data)
		if err != nil { 
			err = errors.New("failed to store data for >" + key + "< with error: " + err.Error())
		}
	}

	return err
}

func (as *AccountStorage) FetchData(client *mautrix.Client, key string, data interface{}) error {
	as.storage_mutex.Lock()
	defer as.storage_mutex.Unlock()
	var err error

	cached_data, exists := as.storage_cache[as.storage_prefix+key]
	if exists {
		err = json.Unmarshal(cached_data, data)
		if err != nil {
			return errors.New("failed to fetch data: "+err.Error())
		}
		return nil
	}
	
	if as.storage_commitupstream == true { //Only attempt to fetch if we dont operate local only
		err = client.GetAccountData(as.storage_prefix+key, &data)
		if err != nil && !errors.Is(err, mautrix.MNotFound) {
			return errors.New("failed to fetch data from upstream for >" + key + "< with error: " + err.Error())
		}

		as.storage_cache[as.storage_prefix+key], err = json.Marshal(data)
		if err != nil {
			return errors.New("failed to cache for fetch for >" + key + "< with data >" + fmt.Sprintf("%+v", data) + "and error: " + err.Error())
		}
	}

	return nil
}
