package global

import (
	"sync"

	"github.com/omatztw/gomatalk/pkg/voice"
)

var (
	VoiceInstances = map[string]*voice.VoiceInstance{}
	Mutex          sync.Mutex
	SpeechSignal   chan voice.SpeechSignal
	// globalMutex sync.Mutex
	// songSignal     chan PkgSong
	// radioSignal    chan PkgRadio
	//ignore            = map[string]bool{}
)