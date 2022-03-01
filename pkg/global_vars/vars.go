package global

import (
	"sync"

	"github.com/omatztw/gomatalk/pkg/db"
	"github.com/omatztw/gomatalk/pkg/voice"
)

var (
	VoiceInstances = map[string]*voice.VoiceInstance{}
	Mutex          sync.Mutex
	SpeechSignal   chan voice.SpeechSignal
	DB             *db.Database
	// globalMutex sync.Mutex
	// songSignal     chan PkgSong
	// radioSignal    chan PkgRadio
	//ignore            = map[string]bool{}
)
