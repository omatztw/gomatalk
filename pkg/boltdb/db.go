package boltdb

import (
	"encoding/json"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/omatztw/gomatalk/pkg/db"
	global "github.com/omatztw/gomatalk/pkg/global_vars"
	"github.com/omatztw/gomatalk/pkg/model"
)

const (
	// filename for db
	dbFileName     string = "data/gomatalk.db"
	rootBucketName string = "GomatailDB"
)

func ListBots(guildID string) (map[string][]string, error) {
	botsList := make(map[string][]string)
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return botsList, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		w := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("config")).Bucket([]byte("bot"))
		c := w.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wavList := []string{}
			err := json.Unmarshal(v, &wavList)
			if err != nil {
				return err
			}
			botsList[string(k)] = wavList
		}
		return nil
	})
	if err != nil {
		return botsList, err
	}
	return botsList, nil
}

func ListWords(guildID string) (map[string]string, error) {
	wordsList := make(map[string]string)
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return wordsList, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		w := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("words"))
		c := w.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wordsList[string(k)] = string(v)
		}
		return nil
	})
	if err != nil {
		return wordsList, err
	}
	return wordsList, nil
}

func ListUser() (map[string]model.UserInfo, error) {
	userList := make(map[string]model.UserInfo)
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return userList, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		w := tx.Bucket([]byte(rootBucketName))
		c := w.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tmpUser model.UserInfo
			err = json.Unmarshal(v, &tmpUser)
			if err != nil {
				continue
			}
			userList[string(k)] = tmpUser
		}
		return nil
	})
	if err != nil {
		return userList, err
	}
	return userList, nil
}

// MIGRATE: bolt to sqlite
func Migrate() error {
	_, err := os.Stat(dbFileName)
	if err != nil {
		// No need to migrate
		return err
	}

	var userList []db.User
	userListDao, _ := ListUser()

	for k, v := range userListDao {
		tmpUser := db.User{
			ID:       k,
			UserInfo: v,
		}
		userList = append(userList, tmpUser)
	}
	global.DB.BulkAddUser(userList)

	guilds, _ := global.DB.ListGuilds()
	for _, guild := range guilds {
		bots, _ := ListBots(guild.ID)
		for k, v := range bots {
			global.DB.AddBot(guild.ID, k, v)
		}
		words, _ := ListWords(guild.ID)
		for k, v := range words {
			global.DB.AddWord(guild.ID, k, v)
		}
	}

	migratedFileName := dbFileName + ".migrated"
	if err := os.Rename(dbFileName, migratedFileName); err != nil {
		return err
	}
	return nil
}
