package util

import (
	"log"
	"math/rand"
	"os"
	"time"
)

func Write(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("FATA: ", err) //ファイルが開けなかったときエラー出力
		return err
	}
	defer file.Close()
	file.Write(([]byte)(content))
	return nil
}

func RandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
