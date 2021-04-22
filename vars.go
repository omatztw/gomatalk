package main

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	dg             *discordgo.Session
	voiceInstances = map[string]*VoiceInstance{}
	mutex          sync.Mutex
	speechSignal   chan SpeechSignal
	globalMutex sync.Mutex
	// songSignal     chan PkgSong
	// radioSignal    chan PkgRadio
	//ignore            = map[string]bool{}
)
