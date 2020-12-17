package main

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Options gomatalk option
type Options struct {
	DiscordToken  string
	DiscordStatus string
	DiscordPrefix string
}

// UserInfo user information for talk
type UserInfo struct {
	Voice     string
	Speed     float64
	Tone      float64
	Intone    float64
	Threshold float64
	Volume    float64
}

type Speech struct {
	Text     string
	UserInfo UserInfo
}

type SpeechSignal struct {
	data Speech
	v    *VoiceInstance
}

type VoiceInstance struct {
	voice      *discordgo.VoiceConnection
	session    *discordgo.Session
	queueMutex sync.Mutex
	voiceMutex sync.Mutex
	nowTalking Speech
	queue      []Speech
	recv       []int16
	guildID    string
	channelID  string
	speaking   bool
	stop       chan bool
}
