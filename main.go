package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/omatztw/gomatalk/pkg/boltdb"
	"github.com/omatztw/gomatalk/pkg/config"
	"github.com/omatztw/gomatalk/pkg/db"
	"github.com/omatztw/gomatalk/pkg/discord"
	global "github.com/omatztw/gomatalk/pkg/global_vars"
)

func WavGC() {
	go func() {
		t := time.NewTicker(30 * time.Minute) // 30分おきに検索
		defer t.Stop()
		for {
			select {
			case <-t.C:
				files, err := walkMatch("/tmp/voice-*.wav")
				if err != nil {
					log.Println("FATA: Error WavGC():", err)
					return
				}
				for _, file := range files {
					info, err := os.Stat(file)
					if err != nil {
						log.Println("FATA: Error WavGC():", err)
						return
					}
					if info.ModTime().Before(time.Now().Add(-time.Minute * 10)) { // 10分前以前に作られたファイルは消去
						log.Println("Garbage WAV found. Deleting...: " + file)
						os.Remove(file)
					}
				}
			}
		}
	}()
}

func walkMatch(pattern string) ([]string, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func main() {
	// runtime.GOMAXPROCS(1)
	filename := flag.String("f", "bot.toml", "Set path for the config file.")
	flag.Parse()
	log.Println("INFO: Opening", *filename)
	err := config.LoadConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}

	err = config.LoadVoiceConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}

	err = config.LoadVoiceVoxConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}

	err = config.LoadVoiceVoxApiConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}

	err = config.LoadAquestalkConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}

	// Hot reload
	config.Watch()

	global.DB = db.NewDatabase("data/gomatalk-sqlite.db")

	// Connecto to Discord
	err = discord.DiscordConnect()
	if err != nil {
		log.Println("FATA: Discord", err)
		return
	}
	boltdb.Migrate()
	WavGC()

	<-make(chan struct{})
}
