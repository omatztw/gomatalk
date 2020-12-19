package main

import (
	"encoding/json"
	"errors"
	"math/rand"
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

		config, err := g.CreateBucketIfNotExists([]byte("config"))
		if err != nil {
			return err
		}

		_, err = config.CreateBucketIfNotExists([]byte("bot"))
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func Addbot(guildID, botID string, wavList []string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("config")).Bucket([]byte("bot"))
		encoded, err := json.Marshal(wavList)
		if err != nil {
			return err
		}
		return b.Put([]byte(botID), encoded)
	})
	return err
}

func DeleteBot(guildID, botID string) error {
	db, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rootBucketName)).Bucket([]byte(guildID)).Bucket([]byte("config")).Bucket([]byte("bot"))
		return b.Delete([]byte(botID))
	})
	return err
}

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

func random(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()*(max-min) + min
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func InitUser(userID string) (UserInfo, error) {

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(voices))

	user := UserInfo{}
	user.Voice = VoiceList()[num]
	user.Speed = random(0.5, 2)
	user.Tone = random(-20, 20)
	user.Intone = random(0, 4)
	user.Threshold = random(0, 1)
	user.Volume = 1.0
	err := PutUser(userID, user)
	if err != nil {
		return user, err
	}

	return user, nil
}
