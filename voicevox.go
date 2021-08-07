package main

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
)

func GenerateAudioQuery(speech Speech) (AudioQuery, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d", vv.baseURL, url.QueryEscape(speech.Text), getIdByName(speech.UserInfo.Voice))

	audioQuery := AudioQuery{}

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
	for _, v := range vv.Voice {
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

	url := fmt.Sprintf("%s/synthesis?speaker=%d", vv.baseURL, getIdByName(speech.UserInfo.Voice))

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
