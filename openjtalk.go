package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	dictDir       string = "/usr/share/open_jtalk/dic"
	sysVoiceDir   string = "/usr/share/open_jtalk/voices"
	localVoiceDir string = "voices"
)

var (
	voices = map[string]string{
		"normal":  fmt.Sprintf("%s/%s", sysVoiceDir, "mei_normal.htsvoice"),
		"happy":   fmt.Sprintf("%s/%s", sysVoiceDir, "mei_happy.htsvoice"),
		"bashful": fmt.Sprintf("%s/%s", sysVoiceDir, "mei_bashful.htsvoice"),
		"angry":   fmt.Sprintf("%s/%s", sysVoiceDir, "mei_angry.htsvoice"),
		"sad":     fmt.Sprintf("%s/%s", sysVoiceDir, "mei_sad.htsvoice"),
		"male":    fmt.Sprintf("%s/%s", sysVoiceDir, "nitech_jp_atr503_m001.htsvoice"),
		"yoe":     fmt.Sprintf("%s/%s", localVoiceDir, "yoe.htsvoice"),
		"taro":    fmt.Sprintf("%s/%s", localVoiceDir, "taro.htsvoice"),
		"ai":      fmt.Sprintf("%s/%s", localVoiceDir, "ai.htsvoice"),
		"ikuru":   fmt.Sprintf("%s/%s", localVoiceDir, "ikuru.htsvoice"),
		"momo":    fmt.Sprintf("%s/%s", localVoiceDir, "momo.htsvoice"),
		"wamea":   fmt.Sprintf("%s/%s", localVoiceDir, "wamea.htsvoice"),
		"akesato": fmt.Sprintf("%s/%s", localVoiceDir, "akesato.htsvoice"),
		"kanata":  fmt.Sprintf("%s/%s", localVoiceDir, "kanata.htsvoice"),
		"row":     fmt.Sprintf("%s/%s", localVoiceDir, "row.htsvoice"),
		"mizuki":  fmt.Sprintf("%s/%s", localVoiceDir, "mizuki.htsvoice"),
	}
)

func VoiceList() []string {
	keys := make([]string, len(voices))
	i := 0
	for k := range voices {
		keys[i] = k
		i++
	}
	return keys
}

func CreateWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())
	textFileName := fmt.Sprintf("/tmp/voice-%d.txt", time.Now().UnixNano())

	write(textFileName, speech.Text)

	defer os.Remove(textFileName)

	cmd := []string{
		"-x", dictDir,
		"-m", voices[speech.UserInfo.Voice],
		"-ow", wavFileName,
		"-r", fmt.Sprintf("%g", speech.UserInfo.Speed),
		"-fm", fmt.Sprintf("%g", speech.UserInfo.Tone),
		"-jf", fmt.Sprintf("%g", speech.UserInfo.Intone),
		"-u", fmt.Sprintf("%g", speech.UserInfo.Threshold),
		"-g", fmt.Sprintf("%g", speech.UserInfo.Volume),
		textFileName,
	}

	run := exec.Command("open_jtalk", cmd...)

	err := run.Run()
	if err != nil {
		log.Println("FATA: Error run():", err)
		return "", err
	}

	return wavFileName, nil
}

func write(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err) //ファイルが開けなかったときエラー出力
		return err
	}
	defer file.Close()
	file.Write(([]byte)(content))
	return nil
}
