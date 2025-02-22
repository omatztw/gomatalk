package voice

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	I := -28 + speech.UserInfo.Volume

	// 音量調整が必要な場合、ffmpegで処理
	adjustedFileName := fmt.Sprintf("/tmp/voice-adjusted-%d.wav", time.Now().UnixNano())
	cmd := exec.Command("ffmpeg", "-i", wavFileName, "-af",
		fmt.Sprintf("loudnorm=I=%f:TP=-1.5:LRA=11", I),
		"-y", adjustedFileName)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("音量正規化に失敗しました: %v", err)
	}

	// 元のファイルを削除
	os.Remove(wavFileName)
	return adjustedFileName, nil
}
