package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/omatztw/gomatalk/pkg/config"
	"github.com/omatztw/gomatalk/pkg/model"
)

func GenerateAudioQuery(speech Speech) (model.AudioQuery, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d", config.Vv.BaseURL, url.QueryEscape(speech.Text), getIdByName(speech.UserInfo.Voice))

	audioQuery := model.AudioQuery{}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte("")),
	)

	if err != nil {
		log.Println("FATA:", err)
		return audioQuery, err
	}

	response, err := client.Do(req)

	if err != nil {
		log.Println("FATA:", err)
		return audioQuery, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Println("FATA:", err)
		return audioQuery, err
	}

	json.Unmarshal(body, &audioQuery)

	// json.NewDecoder(response.Body).Decode(&audioQuery)

	return audioQuery, nil

}

func getIdByName(name string) int {
	for _, v := range config.Vv.Voice {
		if v.Name == name {
			return v.Id
		}
	}
	return 0
}

func CreateVoiceVoxWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/synthesis?speaker=%d", config.Vv.BaseURL, getIdByName(speech.UserInfo.Voice))

	audioQuery, err := GenerateAudioQuery(speech)

	if err != nil {
		return "", err
	}
	audioQuery.SpeedScale = speech.UserInfo.Speed
	// audioQuery.PitchScale = speech.UserInfo.Tone
	audioQuery.IntonationScale = speech.UserInfo.Intone

	body, err := json.Marshal(audioQuery)
	// fmt.Printf("[+] %s\n", string(body))

	if err != nil {
		log.Println("FATA:", err)
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Println("FATA:", err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")

	response, err := client.Do(req)
	if err != nil {
		log.Println("FATA:", err)
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
