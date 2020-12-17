package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

const (
	// filename for db
	dbFileName     string = "data/gomatalk.db"
	rootBucketName string = "GomatailDB"
)

// CreateDB create a database file if it if was not exist
func CreateDB() error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(rootBucketName))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func CreateGuildDB(guildID string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName))

		g, err := b.CreateBucketIfNotExists([]byte(guildID))
		if err != nil {
			return err
		}

		_, err = g.CreateBucketIfNotExists([]byte("words"))
		if err != nil {
			return err
		}

		_, err = g.CreateBucketIfNotExists([]byte("config"))
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func AddWord(guildID, pre, post string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		w := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("words"))
		return w.Put([]byte(pre), []byte(post))
	})
	return err
}

func DeleteWord(guildID, key string) error {

	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		w := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("words"))
		return w.Delete([]byte(key))
	})
	return err
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

func CreateUserDB(userID string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName))

		_, err = b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func PutReplaceWord(guildID, pre, post string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName))
		g := b.Bucket([]byte(guildID))
		if err != nil {
			return err
		}
		return g.Put([]byte(pre), []byte(post))
	})
	return err
}

// PutDB ignore o unignore a test channel
func PutUser(userID string, user UserInfo) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName))
		encoded, err := json.Marshal(user)
		log.Println("ENCODED", string(encoded))
		if err != nil {
			return err
		}
		return b.Put([]byte(userID), encoded)
	})
	return err
}

// GetDB read if a text channel is ignored
func GetUserInfo(userID string) (UserInfo, error) {
	var v []byte
	var user UserInfo
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return user, err
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		v = b.Get([]byte(userID))
		return nil
	})
	if v == nil {
		return user, errors.New("Cannot find user")
	}
	err = json.Unmarshal(v, &user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func InitUser(userID string) (UserInfo, error) {
	user := UserInfo{}
	user.Voice = "normal"
	user.Speed = 1.0
	user.Tone = 0.0
	user.Intone = 1.0
	user.Threshold = 0.5
	user.Volume = 1.0
	// err := CreateUserDB(userID)
	// if err != nil {
	// 	return user, err
	// }
	err := PutUser(userID, user)
	if err != nil {
		return user, err
	}

	return user, nil
}
