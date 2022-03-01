package db

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/omatztw/gomatalk/pkg/model"
	"github.com/omatztw/gomatalk/pkg/voice"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Database struct {
	DBName string
	DB     *gorm.DB
}

func NewDatabase(dbName string) *Database {
	conn := &Database{
		DBName: dbName,
	}
	err := conn.Connect()
	if err != nil {
		panic("Cannot connect to DB")
	}
	// conn.DB.AutoMigrate(&User{}, &Bot{}, &Word{}, &Guild{})
	err = conn.Migrate()
	if err != nil {
		panic("Cannot migrate DB")
	}
	return conn
}

func (c *Database) Connect() error {
	var err error
	c.DB, err = gorm.Open(sqlite.Open(c.DBName), &gorm.Config{})
	if err != nil {
		log.Println("FATA: ", err)
	}
	return err
}

func (c *Database) AddUser(userID string, userInfo model.UserInfo) error {
	user := User{
		ID:       userID,
		UserInfo: userInfo,
	}
	result := c.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&user)
	return result.Error
}

func (c *Database) NewUser(userID string) (User, error) {
	userInfo := MakeRandom()
	user := User{ID: userID, UserInfo: userInfo}
	result := c.DB.Create(&user)
	return user, result.Error
}

func (c *Database) BulkAddUser(userList []User) error {
	result := c.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&userList)
	return result.Error
}

func (c *Database) GetUser(userID string) (User, error) {
	user := User{}
	result := c.DB.Where("id = ?", userID).First(&user)
	return user, result.Error
}

func (c *Database) AddWord(guildID, pre, post string) error {
	word := Word{
		GuildID: guildID,
		Before:  pre,
		After:   post,
	}
	result := c.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&word)
	return result.Error
}

func (c *Database) DeleteWord(guildID, key string) error {
	result := c.DB.Delete(&Word{GuildID: guildID, Before: key})
	return result.Error
}

func (c *Database) ListWords(guildID string) (map[string]string, error) {
	wordsList := make(map[string]string)
	var words []Word
	result := c.DB.Where("guild_id = ?", guildID).Find(&words)
	for _, word := range words {
		wordsList[word.Before] = word.After
	}
	return wordsList, result.Error
}

func (c *Database) AddBot(guildID, botID string, wavList []string) error {
	bot := Bot{
		ID:      botID,
		GuildID: guildID,
		Wav:     strings.Join(wavList, ","),
	}
	result := c.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&bot)
	return result.Error
}

func (c *Database) DeleteBot(guildID, botID string) error {
	result := c.DB.Delete(&Bot{GuildID: guildID, ID: botID})
	return result.Error
}

func (c *Database) ListBots(guildID string) (map[string][]string, error) {
	botList := make(map[string][]string)
	var bots []Bot
	result := c.DB.Where("guild_id = ?", guildID).Find(&bots)
	for _, bot := range bots {
		botList[bot.ID] = strings.Split(bot.Wav, ",")
	}
	return botList, result.Error
}

func (c *Database) CreateGuild(guildID string) error {
	guild := Guild{ID: guildID}
	result := c.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&guild)
	return result.Error
}

func (c *Database) ListGuilds() ([]Guild, error) {
	var guilds []Guild
	result := c.DB.Find(&guilds)
	return guilds, result.Error
}

func (c *Database) Migrate() error {
	sqlDb, err := c.DB.DB()
	if err != nil {
		log.Println("FATA: ", err)
		return err
	}
	driver, err := sqlite3.WithInstance(sqlDb, &sqlite3.Config{})
	if err != nil {
		log.Println("FATA: ", err)
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrates",
		"ql", driver)
	if err != nil {
		log.Println("FATA: ", err)
		return err
	}
	err = m.Up()
	return err

}

func random(min, max float32) float64 {
	rand.Seed(time.Now().UnixNano())
	return float64(rand.Float32()*(max-min) + min)
}

func MakeRandom() model.UserInfo {

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(voice.Voices()))

	user := model.UserInfo{}
	user.Voice = voice.VoiceList()[num]
	user.Speed = random(0.5, 2)
	if voice.IsVoiceRoid(user.Voice) {
		user.Tone = random(0.5, 2)
		user.Intone = random(0, 2)
	} else {
		user.Tone = random(-20, 20)
		user.Intone = random(0, 4)
	}
	user.Threshold = random(0, 1)
	user.AllPass = 0
	user.Volume = 1.0

	return user
}
