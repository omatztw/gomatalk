package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/dgvoice"
)

const ()

// GlobalPlay talk
func GlobalPlay(speechSig chan SpeechSignal) {
	for {
		select {
		case speech := <-speechSig:
			speech.v.PlayQueue(speech.data)
		}
	}
}

func (v *VoiceInstance) PlayQueue(speech Speech) {
	// add song to queue
	v.QueueAdd(speech)
	if v.speaking {
		// the bot is playing
		return
	}
	go func() {
		// v.voiceMutex.Lock()
		// defer v.voiceMutex.Unlock()

		// 複数チャンネルで同時に音声接続するとノイズが発生するため、globalのlockをかける
		// この実装の場合、多くのチャンネルが同時に接続した場合に遅延が発生するのでできればチャネル毎のlockにしたいが。。
		globalMutex.Lock()
		defer globalMutex.Unlock()
		for {
			if len(v.queue) == 0 {
				return
			}
			v.nowTalking = v.QueueGetSpeech()
			v.speaking = true
			defer func() {
				v.speaking = false
			}()
			// v.voice.Speaking(true)

			err := v.Talk(v.nowTalking)
			if err != nil {
				v.Stop(false)
			}

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
		if isVoiceRoid(speech.UserInfo.Voice) {
			fileName, err = CreateVoiceroidWav(speech)
		} else if isVoiceVox(speech.UserInfo.Voice) {
			fileName, err = CreateVoiceVoxWav(speech)
		} else if isAquesTalk(speech.UserInfo.Voice) {
			fileName, err = CreateAquestalkWav(speech)
		} else {
			fileName, err = CreateWav(speech)
		}
		defer os.Remove(fileName)
		if err != nil {
			return err
		}
	}
	c1 := make(chan string, 1)
	go func() {
		dgvoice.PlayAudioFile(v.voice, fileName, v.stop)
		close(c1)
	}()
	select {
	case <-c1:
		return nil
	case <-time.After(30 * time.Second):
		v.Stop(true)
		return nil
	}
}

func (v *VoiceInstance) Stop(force bool) {
	if v.speaking || force {
		v.stop <- true
	}
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ReplaceWords(guildID string, text *string) error {
	wordList, err := ListWords(guildID)
	if err != nil {
		log.Println("ERR: Cannot get word list.")
		return err
	}

	// Replace long word first
	keys := make([]string, 0, len(wordList))
	for k := range wordList {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		*text = strings.Replace(*text, k, wordList[k], -1)
	}

	return nil
}
