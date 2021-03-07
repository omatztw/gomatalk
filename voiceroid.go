package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func CreateVoiceroidWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())

	response, err := http.Get(fmt.Sprintf("%s/api/v1/audiofile?text=%s&name=%s", vo.baseURL, url.QueryEscape(speech.Text), url.QueryEscape(speech.UserInfo.Voice)))
	if err != nil {
        return "", err
    }
	defer response.Body.Close()

	file, err := os.Create(wavFileName)
	if err != nil {
        return "", err
    }
    defer file.Close()
	io.Copy(file, response.Body)
	return wavFileName, nil
}

