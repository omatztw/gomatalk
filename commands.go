package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HelpReporter
func HelpReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'help'")
	help := "```\nコマンド一覧\n" +
		o.DiscordPrefix + "help or " + o.DiscordPrefix + "h  ->  コマンド一覧と簡単な説明を表示.\n" +
		o.DiscordPrefix + "summon or " + o.DiscordPrefix + "s  ->  読み上げを開始.\n" +
		o.DiscordPrefix + "bye or " + o.DiscordPrefix + "b  ->  読み上げを終了.\n" +
		o.DiscordPrefix + "add_word or " + o.DiscordPrefix + "aw  ->  辞書登録. (" + o.DiscordPrefix + "aw 単語 読み" + ")\n" +
		o.DiscordPrefix + "delete_word or " + o.DiscordPrefix + "dw  ->  辞書削除. (" + o.DiscordPrefix + "dw 単語" + ")\n" +
		o.DiscordPrefix + "words_list or " + o.DiscordPrefix + "wl  ->  辞書一覧を表示.\n" +
		o.DiscordPrefix + "add_bot or " + o.DiscordPrefix + "ab  ->  BOTを読み上げ対象に登録. (" + o.DiscordPrefix + "ab <BOT ID> <WAV LIST>" + ")\n" +
		o.DiscordPrefix + "delete_bot or " + o.DiscordPrefix + "db  ->  BOTを読み上げ対象から削除. (" + o.DiscordPrefix + "db <BOT ID>" + ")\n" +
		o.DiscordPrefix + "bots_list or " + o.DiscordPrefix + "bl  ->  読み上げ対象BOTの一覧を表示.\n" +
		o.DiscordPrefix + "status ->  現在の声の設定を表示.\n" +
		o.DiscordPrefix + "update_voice or " + o.DiscordPrefix + "uv  ->  声の設定を変更. (" + o.DiscordPrefix + "uv voice speed tone intone threshold volume" + ")\n" +
		"   voice: 声の種類 [" + strings.Join(VoiceList(), ",") + "]\n" +
		"   speed: 話す速度 範囲(0.5~2.0)(初期値 1.0) \n" +
		"   tone : 声のトーン 範囲(-20~20)(初期値 0.0) \n" +
		"   intone : 声のイントネーション 範囲(0.0~4.0)(初期値 1.0) \n" +
		"   threshold : ブツブツするときとか改善するかも?? 範囲(0.0~1.0)(初期値 0.5) \n" +
		"   volume : 音量（dB） 範囲(-20~20)(初期値 1) \n" +
		o.DiscordPrefix + "stop  ->  読み上げを一時停止.\n```"

	ChMessageSend(m.ChannelID, help)
	//ChMessageSendEmbed(m.ChannelID, "Help", help)
}

// JoinReporter
func JoinReporter(v *VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'join'")
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if voiceChannelID == "" {
		log.Println("ERROR: Voice channel id not found.")
		ChMessageSend(m.ChannelID, "<@"+m.Author.ID+"> まずVCにはいろ( ˘ω˘ )")
		return
	}
	if v != nil {
		log.Println("INFO: Voice Instance already created.")
	} else {
		guildID := SearchGuild(m.ChannelID)
		// create new voice instance
		mutex.Lock()
		v = new(VoiceInstance)
		voiceInstances[guildID] = v
		v.guildID = guildID
		v.session = s
		v.stop = make(chan bool)
		mutex.Unlock()
		//v.InitVoice()
	}
	var err error
	v.channelID = m.ChannelID
	v.voice, err = dg.ChannelVoiceJoin(v.guildID, voiceChannelID, false, false)
	if err != nil {
		v.Stop()
		log.Println("ERROR: Error to join in a voice channel: ", err)
		return
	}
	v.voice.Speaking(false)
	log.Println("INFO: New Voice Instance created")
	ChMessageSend(v.channelID, "おあ")
}

// LeaveReporter
func LeaveReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'leave'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	v.Stop()
	time.Sleep(200 * time.Millisecond)
	v.voice.Disconnect()
	log.Println("INFO: Voice channel destroyed")
	mutex.Lock()
	delete(voiceInstances, v.guildID)
	mutex.Unlock()
	dg.UpdateStatus(0, o.DiscordStatus)
	ChMessageSend(v.channelID, "おつぅ")
}

func ListBotReporter(m *discordgo.MessageCreate) {
	botList, err := ListBots(m.GuildID)
	if err != nil {
		return
	}

	msg := "```\n登録されているBOT一覧\n\n"
	for k, v := range botList {
		name := k
		botUser, err := dg.User(k)
		if err == nil {
			name = botUser.Username
		} else {
			webhook, err := dg.Webhook(k)
			if err == nil {
				name = webhook.Name
			}
		}

		msg += fmt.Sprintf("・BOT: %s(%s)、WAV LIST: %s\n", name, k, strings.Join(v, ","))
	}
	msg += "```"

	ChMessageSend(m.ChannelID, msg)
}

func AddBotReporter(m *discordgo.MessageCreate) {

	commands := splitString(m.Content)
	if len(commands) < 2 {
		HelpReporter(m)
		return
	}
	var username string
	botUser, err := dg.User(commands[1])
	if err != nil {
		webHook, err := dg.Webhook(commands[1])
		if err != nil {
			ChMessageSend(m.ChannelID, fmt.Sprintf("ID「%s」のBOTは見つかりませんでした。", commands[1]))
			return
		}
		username = webHook.Name
	} else {
		username = botUser.Username
	}
	wavList := []string{}
	if len(commands) > 2 {
		wavList = strings.Split(commands[2], ",")
	}
	err = Addbot(m.GuildID, commands[1], wavList)
	if err != nil {
		ChMessageSend(m.ChannelID, fmt.Sprintf("BOT「%s」の登録に失敗しました。", username))
		return
	}
	ChMessageSend(m.ChannelID, fmt.Sprintf("BOT「%s」を読み上げ対象に登録しました。", username))
}

func DeleteBotReporter(m *discordgo.MessageCreate) {

	commands := splitString(m.Content)
	if len(commands) != 2 {
		HelpReporter(m)
		return
	}
	err := DeleteBot(m.GuildID, commands[1])
	if err != nil {
		ChMessageSend(m.ChannelID, fmt.Sprintf("BOT ID「%s」の削除に失敗しました", commands[1]))
		return
	}
	ChMessageSend(m.ChannelID, fmt.Sprintf("BOT ID「%s」を削除しました", commands[1]))
}

func ListWordsReporter(m *discordgo.MessageCreate) {
	wordsList, err := ListWords(m.GuildID)
	if err != nil {
		return
	}

	msg := "```\n登録されている単語一覧\n\n"
	for k, v := range wordsList {
		msg += fmt.Sprintf("・単語: %s、読み: %s\n", k, v)
	}
	msg += "```"

	ChMessageSend(m.ChannelID, msg)
}

func AddWordReporter(m *discordgo.MessageCreate) {

	commands := splitString(m.Content)
	if len(commands) != 3 {
		HelpReporter(m)
		return
	}
	err := AddWord(m.GuildID, commands[1], commands[2])
	if err != nil {
		ChMessageSend(m.ChannelID, fmt.Sprintf("単語「%s」の登録に失敗しました", commands[1]))
		return
	}
	ChMessageSend(m.ChannelID, fmt.Sprintf("単語「%s」を読み「%s」で登録しました", commands[1], commands[2]))
}

func DeleteWordReporter(m *discordgo.MessageCreate) {

	commands := splitString(m.Content)
	if len(commands) != 2 {
		HelpReporter(m)
		return
	}
	err := DeleteWord(m.GuildID, commands[1])
	if err != nil {
		ChMessageSend(m.ChannelID, fmt.Sprintf("単語「%s」の削除に失敗しました", commands[1]))
		return
	}
	ChMessageSend(m.ChannelID, fmt.Sprintf("単語「%s」を削除しました", commands[1]))
}

func splitString(s string) []string {
	// Split string
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = ' ' // space
	fields, err := r.Read()
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return fields
}

func StatusReporter(m *discordgo.MessageCreate) {
	userInfo, err := GetUserInfo(m.Author.ID)
	if err != nil {
		log.Println("ERROR: Cannot get user information.")
		return
	}
	msg := fmt.Sprintf("voice: %s, speed: %g, tone: %g, intone: %g, threshold: %g, volume: %g\n%suv %s %g %g %g %g %g",
		userInfo.Voice,
		userInfo.Speed,
		userInfo.Tone,
		userInfo.Intone,
		userInfo.Threshold,
		userInfo.Volume,
		o.DiscordPrefix,
		userInfo.Voice,
		userInfo.Speed,
		userInfo.Tone,
		userInfo.Intone,
		userInfo.Threshold,
		userInfo.Volume)
	ChMessageSendEmbed(m.ChannelID, msg, "", *m.Author)
}

func SetStatusHandler(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 7 {
		HelpReporter(m)
		return
	}

	keys := make([]string, 0, len(voices))
	for k := range voices {
		keys = append(keys, k)
	}
	voice := commands[1]
	speed := commands[2]
	tone := commands[3]
	intone := commands[4]
	threshold := commands[5]
	volume := commands[6]
	_, ok := voices[voice]
	if !ok {
		log.Println("Not find key", voice)
		HelpReporter(m)
		return
	}
	if err := CheckRange(speed, 0.5, 2.0); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(tone, -20, 20); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(intone, 0, 4); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(threshold, 0, 1); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(volume, -20, 20); err != nil {
		HelpReporter(m)
		return
	}

	userInfo := UserInfo{}
	userInfo.Voice = voice
	userInfo.Speed, _ = strconv.ParseFloat(speed, 64)
	userInfo.Tone, _ = strconv.ParseFloat(tone, 64)
	userInfo.Intone, _ = strconv.ParseFloat(intone, 64)
	userInfo.Threshold, _ = strconv.ParseFloat(threshold, 64)
	userInfo.Volume, _ = strconv.ParseFloat(volume, 64)

	PutUser(m.Author.ID, userInfo)
	StatusReporter(m)
}

func StopReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'stop'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.voice.ChannelID != voiceChannelID {
		return
	}
	v.Stop()
}

func SpeechText(v *VoiceInstance, m *discordgo.MessageCreate) {
	content, err := m.Message.ContentWithMoreMentionsReplaced(v.session)
	if err != nil {
		log.Println("ERROR: Convert Error.")
		return
	}
	// Replace Custom Emoji String
	rep := regexp.MustCompile(`<:([^:]+):\d{18}>`)
	content = rep.ReplaceAllString(content, "$1")

	urlRep := regexp.MustCompile(`https?://([\w-]+\.)+[\w-]+(/[-\w ./?%&=#]*)?`)
	content = urlRep.ReplaceAllString(content, "URL")

	ReplaceWords(v.guildID, &content)

	user, err := GetUserInfo(m.Author.ID)
	if err != nil {
		log.Println("INFO: Cannot Get User info")
		user, err = InitUser(m.Author.ID)
		if err != nil {
			log.Println("ERR: Cannot initialize User")
			return
		}
	}
	botList, _ := ListBots(v.guildID)
	wavFileName := ""
	for k, v := range botList {
		if k == m.Author.ID {
			if len(v) != 0 {
				num := randomInt(0, len(v))
				wavFileName = v[num]
			}
		}
	}
	speech := Speech{content, user, wavFileName}
	speechSig := SpeechSignal{speech, v}
	go func() {
		speechSignal <- speechSig
	}()
	// v.Talk(speech)
}

func CheckRange(val string, min, max float64) error {
	fval, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	if fval < min || max < fval {
		return errors.New("out of range")
	}
	return nil
}
