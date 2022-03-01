package db

import (
	"time"

	"github.com/omatztw/gomatalk/pkg/model"
)

type Guild struct {
	ID        string `gorm:"primaryKey"`
	Bots      []Bot
	Words     []Word
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Bot struct {
	ID        string `gorm:"primaryKey"`
	GuildID   string `gorm:"primaryKey"`
	Wav       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Word struct {
	GuildID   string `gorm:"primaryKey"`
	Before    string `gorm:"primaryKey"`
	After     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID        string         `gorm:"primaryKey"`
	UserInfo  model.UserInfo `gorm:"embedded"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
