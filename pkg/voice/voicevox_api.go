package voice

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/omatztw/gomatalk/pkg/config"
)

func getVoiceIdByName(name string) int {
	for _, v := range config.Va.Voicevox.Voice {
		if v.Name == name {
			return v.Id
		}
	}
	return 0
}

func CreateVoiceVoxApiWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s/?key=%s&speaker=%d&intonationScale=%f&speed=%f&text=%s",
		config.Va.Voicevox.BaseURL,
		config.Va.Voicevox.ApiKey,
		getVoiceIdByName(speech.UserInfo.Voice),
		speech.UserInfo.Intone,
		speech.UserInfo.Speed,
		speech.Text,
	)
	req, err := http.NewRequest(
		"GET",
		url,
		bytes.NewBuffer([]byte("")),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "*/*")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	file, err := os.Create(wavFileName)
	if err != nil {
		log.Println("FATA:", err)
		return "", err
	}

	defer file.Close()
	io.Copy(file, response.Body)

	return wavFileName, nil
}
