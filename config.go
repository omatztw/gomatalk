package main

import (
	"errors"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var o = &Options{}
var vo = &VoiceRoidConfig{}
var vv = &VoicevoxConfig{}
var aq = &AquestalkConfig{}

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
		return err
	}
	if o.DiscordToken = viper.GetString("discord.token"); o.DiscordToken == "" {
		return errors.New("'token' must be present in config file")
	}
	if o.DiscordStatus = viper.GetString("discord.status"); o.DiscordStatus == "" {
		return errors.New("'status' must be present in config file")
	}
	if o.DiscordPrefix = viper.GetString("discord.prefix"); o.DiscordPrefix == "" {
		return errors.New("'prefix' must be present in config file")
	}
	o.DiscordNumShard = viper.GetInt("discord.shardCount")
	o.DiscordShardID = viper.GetInt("discord.shardID")
	o.Debug = viper.GetBool("discord.debug")
	return nil
}

func LoadVoiceConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	vo.baseURL = viper.GetString("voiceroid.baseURL")
	var voiceRoid []VoiceRoid
	viper.UnmarshalKey("voiceroid.voice", &voiceRoid)
	vo.Voice = voiceRoid
	return nil
}

func LoadVoiceVoxConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	vv.baseURL = viper.GetString("voicevox.baseURL")
	var voiceVox []VoiceVox
	viper.UnmarshalKey("voicevox.voice", &voiceVox)
	vv.Voice = voiceVox
	return nil
}

func LoadAquestalkConfig(filename string) (err error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	aq.ExePath = viper.GetString("aquestalk.exePath")
	var aquestalk []Aquestalk
	viper.UnmarshalKey("aquestalk.voice", &aquestalk)
	aq.Voice = aquestalk
	return nil
}
