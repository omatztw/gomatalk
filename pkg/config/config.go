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
	LoadVoiceConfig(e.Name)
	LoadVoiceVoxConfig(e.Name)
	LoadAquestalkConfig(e.Name)
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
	err = viper.Unmarshal(&O)
	if err != nil {
		return errors.New("cannot load config")
	}
	if O.Discord.Token == "" {
		return errors.New("'token' must be present in config file")
	}
	if O.Discord.Status == "" {
		return errors.New("'status' must be present in config file")
	}
	if O.Discord.Prefix == "" {
		return errors.New("'prefix' must be present in config file")
	}
	return nil
}

func LoadVoiceConfig(filename string) (err error) {
	Vo = &model.VoiceRoidConfig{}
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&Vo)
	if err != nil {
		return errors.New("cannot load config")
	}
	return nil
}

func LoadVoiceVoxConfig(filename string) (err error) {
	Vv = &model.VoicevoxConfig{}
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&Vv)
	if err != nil {
		return errors.New("cannot load config")
	}
	return nil
}

func LoadAquestalkConfig(filename string) (err error) {
	Aq = &model.AquestalkConfig{}
	viper.SetConfigType("toml")
	viper.SetConfigFile(filename)

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&Aq)
	if err != nil {
		return errors.New("cannot load config")
	}
	return nil
}
