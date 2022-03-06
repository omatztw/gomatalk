package model

// Options gomatalk option
type Options struct {
	Discord struct {
		Token    string `mapstructure:"token"`
		Status   string `mapstructure:"status"`
		Prefix   string `mapstructure:"prefix"`
		NumShard int    `mapstructure:"shardCount"`
		ShardID  int    `mapstructure:"shardID"`
		Debug    bool   `mapstructure:"debug"`
		Secret   string `mapstructure:"secret"`
	} `mapstructure:"discord"`
	Greeting map[string]string `mapstructure:"greeting"`
	ErrorMsg map[string]string `mapstructure:"errorMsg"`
}

// UserInfo user information for talk
type UserInfo struct {
	Voice     string
	Speed     float64
	Tone      float64
	Intone    float64
	Threshold float64
	AllPass   float64
	Volume    float64
}

type VoiceRoidConfig struct {
	Voiceroid struct {
		BaseURL string      `mapstructure:"baseURL"`
		Voice   []VoiceRoid `mapstructure:"voice"`
	} `mapstructure:"voiceroid"`
}

type VoiceRoid struct {
	Name string
}

type VoicevoxConfig struct {
	Voicevox struct {
		BaseURL string     `mapstructure:"baseURL"`
		Voice   []VoiceVox `mapstructure:"voice"`
	} `mapstructure:"voicevox"`
}

type VoiceVox struct {
	Name string
	Id   int
}

type AudioQuery struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speedScale"`
	PitchScale         float64        `json:"pitchScale"`
	IntonationScale    float64        `json:"intonationScale"`
	VolumeScale        float64        `json:"volumeScale"`
	PrePhonemeLength   float64        `json:"prePhonemeLength"`
	PostPhonemeLength  float64        `json:"postPhonemeLength"`
	OutputSamplingRate int            `json:"outputSamplingRate"`
	OutputStereo       bool           `json:"outputStereo"`
}

type Mora struct {
	Text            string  `json:"text,omitempty"`
	Consonant       string  `json:"consonant,omitempty"`
	ConsonantLength float64 `json:"consonant_length,omitempty"`
	Vowel           string  `json:"vowel,omitempty"`
	VowelLength     float64 `json:"vowel_length,omitempty"`
	Pitch           float64 `json:"pitch"`
}

type AccentPhrase struct {
	Moras  []Mora `json:"moras"`
	Accent int    `json:"accent"`
}

type AquestalkConfig struct {
	Aquestalk struct {
		ExePath string      `mapstructure:"exePath"`
		Voice   []Aquestalk `mapstructure:"voice"`
	} `mapstructure:"aquestalk"`
}

type Aquestalk struct {
	Name string
}
