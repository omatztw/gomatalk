package main

import (
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
		o.DiscordPrefix + "random or " + o.DiscordPrefix + "r  ->  自分の声をﾗﾝﾀﾞﾑで変更する.\n" +
		o.DiscordPrefix + "status ->  現在の声の設定を表示.\n" +
		o.DiscordPrefix + "update_voice or " + o.DiscordPrefix + "uv  ->  声の設定を変更. (" + o.DiscordPrefix + "uv voice speed tone intone threshold volume" + ")\n" +
		"   voice: 声の種類 - " + strings.Join(VoiceList(), "\n                  - ") + "\n" +
		"   speed: 話す速度 範囲(0.5~2.0) \n" +
		"   tone : 声のトーン 範囲(-20~20) [VOICEROIDは 0.5 ~ 2] \n" +
		"   intone : 声のイントネーション 範囲(0.0~4.0)(初期値 1.0) [VOICEROIDは 0 ~ 2] \n" +
		"   threshold : ブツブツするときとか改善するかも?? 範囲(0.0~1.0)(初期値 0.5) \n" +
		"   allpass : よくわからん 範囲(0 - 1.0) (0はauto)  \n" +
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
		v.Stop(false)
		log.Println("ERROR: Error to join in a voice channel: ", err)
		return
	}
	if o.Debug {
		v.voice.LogLevel = discordgo.LogDebug
	}
	// v.voice.Speaking(false)
	log.Println("INFO: New Voice Instance created")
	botUser, _ := dg.User("@me")
	channel, err := dg.Channel(m.ChannelID)
	if err == nil {
		nickname := botUser.Username + "(" + channel.Name + ")"
		updateNickName(v, nickname)
	}
	ChMessageSend(v.channelID, "おあ")
}

func updateNickName(v *VoiceInstance, nickname string) {
	v.session.GuildMemberNickname(v.guildID, "@me", nickname)
}

// LeaveReporter
func LeaveReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'leave'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	time.Sleep(200 * time.Millisecond)
	v.voice.Disconnect()
	log.Println("INFO: Voice channel destroyed")
	mutex.Lock()
	delete(voiceInstances, v.guildID)
	mutex.Unlock()
	dg.UpdateGameStatus(0, o.DiscordStatus)
	updateNickName(v, "")
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
	// Split string with space
	re := regexp.MustCompile(`['"](\s*[^'"]+)\s*['"]|(\S+)`)
	result := re.FindAllStringSubmatch(s, -1)
	var fields []string
	for _, val := range result {
		if val[1] != "" {
			fields = append(fields, val[1])
		} else {
			fields = append(fields, val[0])
		}
	}
	return fields
}

func StatusReporter(m *discordgo.MessageCreate) {
	statusReporterInternal(m.Author.ID, m)
}

func StatusReporterForOther(userID string, m *discordgo.MessageCreate) {
	statusReporterInternal(userID, m)
}

func statusReporterInternal(userID string, m *discordgo.MessageCreate) {
	user, err := dg.User(userID)
	if err != nil {
		webHook, err := dg.Webhook(userID)
		if err != nil {
			log.Println("ERROR: Cannot find user information.")
			return
		}
		user = webHook.User
	}
	userInfo, err := GetUserInfo(userID)
	if err != nil {
		log.Println("ERROR: Cannot get user information.")
		return
	}
	msg := fmt.Sprintf("voice: %s, speed: %.1f, tone: %.1f, intone: %.1f, threshold: %.1f, allpass: %.1f, volume: %.1f\n%suv %s %.1f %.1f %.1f %.1f %.1f %.1f",
		userInfo.Voice,
		userInfo.Speed,
		userInfo.Tone,
		userInfo.Intone,
		userInfo.Threshold,
		userInfo.AllPass,
		userInfo.Volume,
		o.DiscordPrefix,
		userInfo.Voice,
		userInfo.Speed,
		userInfo.Tone,
		userInfo.Intone,
		userInfo.Threshold,
		userInfo.AllPass,
		userInfo.Volume)
	ChMessageSendEmbed(m.ChannelID, msg, "", *user)
}

func MakeRandomForOther(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 2 {
		HelpReporter(m)
		return
	}
	userID := commands[1]
	user, _ := dg.User(userID)
	if !user.Bot {
		ChMessageSend(m.ChannelID, "声変えられるのはBotだけ( ˘ω˘ )")
		return
	}
	makeRandomHandlerInternal(userID, m)
}

func MakeRandomHandler(m *discordgo.MessageCreate) {
	makeRandomHandlerInternal(m.Author.ID, m)
}

func makeRandomHandlerInternal(userID string, m *discordgo.MessageCreate) {
	user := MakeRandom()
	PutUser(userID, user)
	statusReporterInternal(userID, m)
}

func setStatusHandlerInternal(userID string, userInfo UserInfo, m *discordgo.MessageCreate) {
	_, ok := Voices()[userInfo.Voice]
	if !ok {
		log.Println("Not find key", userInfo.Voice)
		HelpReporter(m)
		return
	}
	if err := CheckRange(userInfo.Speed, 0.5, 2.0); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(userInfo.Tone, -20, 20); err != nil {
		HelpReporter(m)
		return
	}

	if isVoiceRoid(userInfo.Voice) {
		if err := CheckRange(userInfo.Tone, 0.5, 2); err != nil {
			HelpReporter(m)
			return
		}
	}

	if err := CheckRange(userInfo.Intone, 0, 4); err != nil {
		HelpReporter(m)
		return
	}

	if isVoiceRoid(userInfo.Voice) {
		if err := CheckRange(userInfo.Intone, 0, 2); err != nil {
			HelpReporter(m)
			return
		}
	}

	if err := CheckRange(userInfo.Threshold, 0, 1); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(userInfo.Volume, -20, 20); err != nil {
		HelpReporter(m)
		return
	}
	if err := CheckRange(userInfo.AllPass, 0, 1); err != nil {
		HelpReporter(m)
		return
	}
	PutUser(userID, userInfo)
	statusReporterInternal(userID, m)

}

func SetStatusForOtherHandler(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 9 {
		HelpReporter(m)
		return
	}
	userID := commands[1]

	user, err := dg.User(userID)
	if err != nil {
		_, err := dg.Webhook(userID)
		if err != nil {
			ChMessageSend(m.ChannelID, fmt.Sprintf("ID「%s」のBOTは見つかりませんでした。", userID))
			return
		}
	} else {
		if !user.Bot {
			ChMessageSend(m.ChannelID, "声変えられるのはBotだけ( ˘ω˘ )")
			return
		}
	}

	voice := commands[2]
	speed := commands[3]
	tone := commands[4]
	intone := commands[5]
	threshold := commands[6]
	allpass := commands[7]
	volume := commands[8]

	userInfo := UserInfo{}
	userInfo.Voice = voice
	userInfo.Speed, _ = strconv.ParseFloat(speed, 32)
	userInfo.Tone, _ = strconv.ParseFloat(tone, 32)
	userInfo.Intone, _ = strconv.ParseFloat(intone, 32)
	userInfo.Threshold, _ = strconv.ParseFloat(threshold, 32)
	userInfo.AllPass, _ = strconv.ParseFloat(allpass, 32)
	userInfo.Volume, _ = strconv.ParseFloat(volume, 32)

	setStatusHandlerInternal(userID, userInfo, m)
}

func SetStatusHandler(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 8 {
		HelpReporter(m)
		return
	}

	voice := commands[1]
	speed := commands[2]
	tone := commands[3]
	intone := commands[4]
	threshold := commands[5]
	allpass := commands[6]
	volume := commands[7]

	userInfo := UserInfo{}
	userInfo.Voice = voice
	userInfo.Speed, _ = strconv.ParseFloat(speed, 32)
	userInfo.Tone, _ = strconv.ParseFloat(tone, 32)
	userInfo.Intone, _ = strconv.ParseFloat(intone, 32)
	userInfo.Threshold, _ = strconv.ParseFloat(threshold, 32)
	userInfo.AllPass, _ = strconv.ParseFloat(allpass, 32)
	userInfo.Volume, _ = strconv.ParseFloat(volume, 32)

	setStatusHandlerInternal(m.Author.ID, userInfo, m)
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
	v.Stop(true)
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

	urlRep := regexp.MustCompile(`https?://[\w!\?/\+\-_~=;\.,\*&@#\$%\(\)'\[\]]+`)
	content = urlRep.ReplaceAllString(content, "URL")

	slashCommand := regexp.MustCompile(`</([^:]+):\d{18}>`)
	content = slashCommand.ReplaceAllString(content, "$1")

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

func CheckRange(val float64, min, max float64) error {
	if val < min || max < val {
		return errors.New("out of range")
	}
	return nil
}
