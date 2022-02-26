package play

import (
	"log"
	"os"
	"sort"
	"strings"

	"github.com/omatztw/gomatalk/pkg/db"
	"github.com/omatztw/gomatalk/pkg/voice"
)

const ()

// GlobalPlay talk
func GlobalPlay(speechSig chan voice.SpeechSignal) {
	for {
		select {
		case speech := <-speechSig:
			speech.V.PlayQueue(speech.Data)
		}
	}
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ReplaceWords(guildID string, text *string) error {
	wordList, err := db.ListWords(guildID)
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
