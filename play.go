package main

import (
	"log"
	"os"
	"strings"

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
		v.voiceMutex.Lock()
		defer v.voiceMutex.Unlock()
		for {
			if len(v.queue) == 0 {
				return
			}
			v.nowTalking = v.QueueGetSpeech()
			v.speaking = true
			v.voice.Speaking(true)

			v.Talk(v.nowTalking)

			v.QueueRemoveFisrt()
			v.speaking = false
			v.voice.Speaking(false)
		}
	}()
}

func (v *VoiceInstance) Talk(speech Speech) error {
	ReplaceWords(v.guildID, &speech.Text)
	fileName, err := CreateWav(speech)
	defer os.Remove(fileName)
	if err != nil {
		return err
	}
	dgvoice.PlayAudioFile(v.voice, fileName, v.stop)
	return nil
}

func (v *VoiceInstance) Stop() {
	v.stop <- true
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

	for k, v := range wordList {
		*text = strings.Replace(*text, k, v, -1)
	}
	return nil
}
