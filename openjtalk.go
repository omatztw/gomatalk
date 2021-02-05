package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	}
)

func VoiceList() []string {
	ans := Voices()
	keys := make([]string, len(ans))
	i := 0
	for k := range ans {
		keys[i] = k
		i++
	}
	return keys
}

func Voices() map[string]string {
	return merge(voices, LocalVoiceList())
}

func LocalVoiceList() map[string]string {
	pattern := localVoiceDir + "/*.htsvoice"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return make(map[string]string)
	}
	list := map[string]string{}
	for _, file := range files {
		_, filename := filepath.Split(file)
		filename = filename[0 : len(filename)-len(".htsvoice")]
		list[filename] = file
	}
	return list
}

func merge(m1, m2 map[string]string) map[string]string {
	ans := map[string]string{}

	for k, v := range m1 {
		ans[k] = v
	}
	for k, v := range m2 {
		ans[k] = v
	}
	return (ans)
}

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
						log.Println("FATA: Error DeleteWav():", err)
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

func CreateWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())
	textFileName := fmt.Sprintf("/tmp/voice-%d.txt", time.Now().UnixNano())

	write(textFileName, speech.Text)

	defer os.Remove(textFileName)

	cmd := []string{
		"-x", dictDir,
		"-m", Voices()[speech.UserInfo.Voice],
		"-ow", wavFileName,
		"-r", fmt.Sprintf("%g", speech.UserInfo.Speed),
		"-fm", fmt.Sprintf("%g", speech.UserInfo.Tone),
		"-jf", fmt.Sprintf("%g", speech.UserInfo.Intone),
		"-u", fmt.Sprintf("%g", speech.UserInfo.Threshold),
		"-g", fmt.Sprintf("%g", speech.UserInfo.Volume),
	}

	if speech.UserInfo.AllPass > 0 {
		cmd = append(cmd, "-a")
		cmd = append(cmd, fmt.Sprintf("%g", speech.UserInfo.AllPass))
	}

	cmd = append(cmd, textFileName)

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
