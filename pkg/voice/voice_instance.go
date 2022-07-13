package voice

import (
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/omatztw/dgvoice"
	"github.com/omatztw/gomatalk/pkg/model"
)

type VoiceInstance struct {
	sync.Mutex
	Voice      *discordgo.VoiceConnection
	Session    *discordgo.Session
	QueueMutex sync.Mutex
	VoiceMutex sync.Mutex
	NowTalking Speech
	Queue      []Speech
	Recv       []int16
	GuildID    string
	ChannelID  string
	Speaking   bool
	Stop       chan bool
}

type SpeechSignal struct {
	Data Speech
	V    *VoiceInstance
}

type Speech struct {
	Text     string
	UserInfo model.UserInfo
	WavFile  string
}

func (v *VoiceInstance) PlayQueue(speech Speech) {
	// add song to queue
	v.QueueAdd(speech)
	if v.Speaking {
		// the bot is playing
		return
	}
	go func() {
		// 同一チャンネルで同時に読み上げるのを防ぐ。別のサーバーには影響しないようにしたい。
		v.VoiceMutex.Lock()
		defer v.VoiceMutex.Unlock()

		for {
			if len(v.Queue) == 0 {
				return
			}
			v.NowTalking = v.QueueGetSpeech()
			v.Speaking = true
			defer func() {
				v.Speaking = false
			}()
			// v.voice.Speaking(true)

			v.Talk(v.NowTalking)

			v.QueueRemoveFisrt()
			// v.voice.Speaking(false)
		}
	}()
}

func (v *VoiceInstance) Talk(speech Speech) error {
	var fileName string
	var err error
	if speech.WavFile != "" {
		fileName = "wav/" + speech.WavFile
	} else {
		if IsVoiceRoid(speech.UserInfo.Voice) {
			fileName, err = CreateVoiceroidWav(speech)
		} else if IsVoiceVox(speech.UserInfo.Voice) {
			fileName, err = CreateVoiceVoxWav(speech)
		} else if IsAquesTalk(speech.UserInfo.Voice) {
			fileName, err = CreateAquestalkWav(speech)
		} else {
			fileName, err = CreateWav(speech)
		}
		if err != nil {
			// VOICEROIDやVOICEBOXが起動していない場合に通常音声で再生する
			fallbackSpeech := Speech{
				Text: speech.Text,
				UserInfo: model.UserInfo{
					Voice:     "normal",
					Speed:     1.3,
					Tone:      1,
					Intone:    0,
					Threshold: 0.5,
					AllPass:   0,
					Volume:    1,
				},
				WavFile: speech.WavFile,
			}
			fileName, err = CreateWav(fallbackSpeech)
		}
		defer os.Remove(fileName)
		if err != nil {
			return err
		}
	}
	c1 := make(chan string, 1)
	go func() {
		dgvoice.PlayAudioFile(v.Voice, fileName, v.Stop)
		close(c1)
	}()
	select {
	case <-c1:
		return nil
	case <-time.After(30 * time.Second):
		v.StopTalking()
		return nil
	}
}

func (v *VoiceInstance) StopTalking() {
	if v.Speaking {
		v.Stop <- true
	}
}

// QueueGetSong
func (v *VoiceInstance) QueueGetSpeech() (speech Speech) {
	v.QueueMutex.Lock()
	defer v.QueueMutex.Unlock()
	if len(v.Queue) != 0 {
		return v.Queue[0]
	}
	return
}

// QueueAdd
func (v *VoiceInstance) QueueAdd(speech Speech) {
	v.QueueMutex.Lock()
	defer v.QueueMutex.Unlock()
	v.Queue = append(v.Queue, speech)
}

// QueueClean
func (v *VoiceInstance) QueueClean() {
	v.QueueMutex.Lock()
	defer v.QueueMutex.Unlock()
	v.Queue = []Speech{}
}

// QueueRemoveFirst
func (v *VoiceInstance) QueueRemoveFisrt() {
	v.QueueMutex.Lock()
	defer v.QueueMutex.Unlock()
	if len(v.Queue) != 0 {
		v.Queue = v.Queue[1:]
	}
}
