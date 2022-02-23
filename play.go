package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/omatztw/dgvoice"
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
		// 同一チャンネルで同時に読み上げるのを防ぐ。別のサーバーには影響しないようにしたい。
		v.voiceMutex.Lock()
		defer v.voiceMutex.Unlock()

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

			v.Talk(v.nowTalking)

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
		v.Stop()
		return nil
	}
}

func (v *VoiceInstance) Stop() {
	if v.speaking {
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
