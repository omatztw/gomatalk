package voice

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/omatztw/gomatalk/pkg/config"
	"github.com/omatztw/gomatalk/pkg/util"
)

func CreateAquestalkWav(speech Speech) (string, error) {
	wavFileName := fmt.Sprintf("/tmp/voice-%d.wav", time.Now().UnixNano())
	textFileName := fmt.Sprintf("/tmp/voice-%d.txt", time.Now().UnixNano())

	util.Write(textFileName, speech.Text)

	defer os.Remove(textFileName)

	cmd := []string{
		"-o", wavFileName,
		"-f", textFileName,
		"-s", fmt.Sprintf("%g", speech.UserInfo.Speed*100),
		"-g", fmt.Sprintf("%g", (speech.UserInfo.Volume+20)*2.5),
	}

	run := exec.Command(config.Aq.ExePath, cmd...)

	err := run.Run()
	if err != nil {
		log.Println("FATA: Error run():", err)
		return "", err
	}

	return wavFileName, nil
}
