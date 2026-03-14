package voice

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	disgovoice "github.com/disgoorg/disgo/voice"
	"github.com/omatztw/gomatalk/pkg/model"
	"layeh.com/gopus"
)

type VoiceInstance struct {
	sync.Mutex
	Conn           disgovoice.Conn
	Session        *discordgo.Session
	QueueMutex     sync.Mutex
	VoiceMutex     sync.Mutex
	NowTalking     Speech
	Queue          []Speech
	Recv           []int16
	GuildID        string
	ChannelID      string
	VoiceChannelID string
	Speaking       bool
	Stop           chan bool
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
		} else if IsVoiceVoxApi(speech.UserInfo.Voice) {
			fileName, err = CreateVoiceVoxApiWav(speech)
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
	c1 := make(chan error, 1)
	go func() {
		c1 <- playAudioFile(v.Conn, fileName, v.Stop)
	}()
	select {
	case err := <-c1:
		return err
	case <-time.After(30 * time.Second):
		v.StopTalking()
		<-c1
		return nil
	}
}

func (v *VoiceInstance) StopTalking() {
	if v.Speaking {
		select {
		case v.Stop <- true:
		default:
		}
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

const (
	audioChannels  int = 2
	audioFrameRate int = 48000
	audioFrameSize int = 960
	audioMaxBytes  int = (audioFrameSize * 2) * 2
)

func playAudioFile(conn disgovoice.Conn, filename string, stop <-chan bool) error {
	if conn == nil {
		return errors.New("voice connection is nil")
	}

	drainStop(stop)

	cmd := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(audioFrameRate), "-ac", strconv.Itoa(audioChannels), "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	reader := bufio.NewReaderSize(stdout, 16384)

	if err = cmd.Start(); err != nil {
		return err
	}
	defer cmd.Process.Kill()

	encoder, err := gopus.NewEncoder(audioFrameRate, audioChannels, gopus.Audio)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_ = conn.SetSpeaking(ctx, disgovoice.SpeakingFlagMicrophone)
	cancel()
	defer func() {
		for i := 0; i < 5; i++ {
			_, _ = conn.UDP().Write(disgovoice.SilenceAudioFrame)
		}
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = conn.SetSpeaking(stopCtx, disgovoice.SpeakingFlagNone)
		stopCancel()
	}()

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		pcm := make([]int16, audioFrameSize*audioChannels)
		err = binary.Read(reader, binary.LittleEndian, &pcm)
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil
		}
		if err != nil {
			return err
		}

		frame, err := encoder.Encode(pcm, audioFrameSize, audioMaxBytes)
		if err != nil {
			return err
		}
		if _, err = conn.UDP().Write(frame); err != nil {
			return err
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func drainStop(stop <-chan bool) {
	for {
		select {
		case <-stop:
		default:
			return
		}
	}
}
