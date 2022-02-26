package config

import (
	"errors"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/omatztw/gomatalk/pkg/model"
	"github.com/spf13/viper"
)

var O = &model.Options{}
var Vo = &model.VoiceRoidConfig{}
var Vv = &model.VoicevoxConfig{}
var Aq = &model.AquestalkConfig{}

// Watch hot reload
func Watch() {
	// Hot reload
	viper.WatchConfig()
	viper.OnConfigChange(Reload)
}

// Reload reload conf
func Reload(e fsnotify.Event) {
	log.Println("INFO: The config file changed:", e.Name)
	LoadConfig(e.Name)
	//StopStream()
}

// LoadConfig load conf from file
func LoadConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)
	//viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Println("HOGE")
		return err
	}
	if O.DiscordToken = viper.GetString("discord.token"); O.DiscordToken == "" {
		return errors.New("'token' must be present in config file")
	}
	if O.DiscordStatus = viper.GetString("discord.status"); O.DiscordStatus == "" {
		return errors.New("'status' must be present in config file")
	}
	if O.DiscordPrefix = viper.GetString("discord.prefix"); O.DiscordPrefix == "" {
		return errors.New("'prefix' must be present in config file")
	}
	O.DiscordNumShard = viper.GetInt("discord.shardCount")
	O.DiscordShardID = viper.GetInt("discord.shardID")
	O.Debug = viper.GetBool("discord.debug")
	O.Secret = viper.GetString("discord.secret")
	return nil
}

func LoadVoiceConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	Vo.BaseURL = viper.GetString("voiceroid.baseURL")
	var voiceRoid []model.VoiceRoid
	viper.UnmarshalKey("voiceroid.voice", &voiceRoid)
	Vo.Voice = voiceRoid
	return nil
}

func LoadVoiceVoxConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	Vv.BaseURL = viper.GetString("voicevox.baseURL")
	var voiceVox []model.VoiceVox
	viper.UnmarshalKey("voicevox.voice", &voiceVox)
	Vv.Voice = voiceVox
	return nil
}

func LoadAquestalkConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	Aq.ExePath = viper.GetString("aquestalk.exePath")
	var aquestalk []model.Aquestalk
	viper.UnmarshalKey("aquestalk.voice", &aquestalk)
	Aq.Voice = aquestalk
	return nil
}
