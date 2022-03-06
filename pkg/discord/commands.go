package discord

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/omatztw/gomatalk/pkg/config"
	"github.com/omatztw/gomatalk/pkg/db"
	global "github.com/omatztw/gomatalk/pkg/global_vars"
	"github.com/omatztw/gomatalk/pkg/model"
	"github.com/omatztw/gomatalk/pkg/play"
	"github.com/omatztw/gomatalk/pkg/util"
	"github.com/omatztw/gomatalk/pkg/voice"
)

// HelpReporter
func HelpReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'help'")
	help := "コマンド一覧\n" +
		config.O.Discord.Prefix + "help or " + config.O.Discord.Prefix + "h  ->  コマンド一覧と簡単な説明を表示.\n" +
		config.O.Discord.Prefix + "summon or " + config.O.Discord.Prefix + "s  ->  読み上げを開始.\n" +
		config.O.Discord.Prefix + "bye or " + config.O.Discord.Prefix + "b  ->  読み上げを終了.\n" +
		config.O.Discord.Prefix + "add_word or " + config.O.Discord.Prefix + "aw  ->  辞書登録. (" + config.O.Discord.Prefix + "aw 単語 読み" + ")\n" +
		config.O.Discord.Prefix + "delete_word or " + config.O.Discord.Prefix + "dw  ->  辞書削除. (" + config.O.Discord.Prefix + "dw 単語" + ")\n" +
		config.O.Discord.Prefix + "words_list or " + config.O.Discord.Prefix + "wl  ->  辞書一覧を表示.\n" +
		config.O.Discord.Prefix + "add_bot or " + config.O.Discord.Prefix + "ab  ->  BOTを読み上げ対象に登録. (" + config.O.Discord.Prefix + "ab <BOT ID> <WAV LIST>" + ")\n" +
		config.O.Discord.Prefix + "delete_bot or " + config.O.Discord.Prefix + "db  ->  BOTを読み上げ対象から削除. (" + config.O.Discord.Prefix + "db <BOT ID>" + ")\n" +
		config.O.Discord.Prefix + "bots_list or " + config.O.Discord.Prefix + "bl  ->  読み上げ対象BOTの一覧を表示.\n" +
		config.O.Discord.Prefix + "random or " + config.O.Discord.Prefix + "r  ->  自分の声をﾗﾝﾀﾞﾑで変更する.\n" +
		config.O.Discord.Prefix + "status ->  現在の声の設定を表示.\n" +
		config.O.Discord.Prefix + "update_voice or " + config.O.Discord.Prefix + "uv  ->  声の設定を変更. (" + config.O.Discord.Prefix + "uv voice speed tone intone threshold volume" + ")\n" +
		"   voice: 声の種類 - " + strings.Join(voice.VoiceList(), "\n                  - ") + "\n" +
		"   speed: 話す速度 範囲(0.5~2.0) \n" +
		"   tone : 声のトーン 範囲(-20~20) [VOICEROIDは 0.5 ~ 2] \n" +
		"   intone : 声のイントネーション 範囲(0.0~4.0)(初期値 1.0) [VOICEROIDは 0 ~ 2] \n" +
		"   threshold : ブツブツするときとか改善するかも?? 範囲(0.0~1.0)(初期値 0.5) \n" +
		"   allpass : よくわからん 範囲(0 - 1.0) (0はauto)  \n" +
		"   volume : 音量（dB） 範囲(-20~20)(初期値 1) \n" +
		config.O.Discord.Prefix + "stop  ->  読み上げを一時停止."
	log.Println("", m.ChannelID)
	ChFileSend(m.ChannelID, "help.txt", help)
	// ChMessageSend(m.ChannelID, help)
	//ChMessageSendEmbed(m.ChannelID, "Help", help)
}

// JoinReporter
func JoinReporter(v *voice.VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'join'")
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if voiceChannelID == "" {
		log.Println("ERROR: Voice channel id not found.")
		ChMessageSend(m.ChannelID, "<@"+m.Author.ID+"> "+config.O.ErrorMsg["novc"])
		return
	}
	already := false
	if v != nil {
		log.Println("INFO: A voice instance is already created.")
		if v.ChannelID == m.ChannelID {
			already = true
		}
	} else {
		log.Println("INFO: New Voice Instance created")
		guildID := SearchGuild(m.ChannelID)
		// create new voice instance
		global.Mutex.Lock()
		v = new(voice.VoiceInstance)
		global.VoiceInstances[guildID] = v
		v.GuildID = guildID
		v.Session = s
		v.Stop = make(chan bool, 1)
		global.Mutex.Unlock()
		//v.InitVoice()
	}
	var err error
	v.ChannelID = m.ChannelID
	v.Voice, err = Dg.ChannelVoiceJoin(v.GuildID, voiceChannelID, false, false)
	if err != nil {
		v.StopTalking()
		log.Println("ERROR: Error to join in a voice channel: ", err)
		return
	}
	if config.O.Discord.Debug {
		v.Voice.LogLevel = discordgo.LogDebug
	}
	// v.Voice.Speaking(false)

	botUser, _ := Dg.User("@me")
	channel, err := Dg.Channel(m.ChannelID)
	if err == nil {
		nickname := botUser.Username + "(" + channel.Name + ")"
		updateNickName(v, nickname)
	}
	if !already {
		ChMessageSend(v.ChannelID, config.O.Greeting["join"])
	}
}

func updateNickName(v *voice.VoiceInstance, nickname string) {
	v.Session.GuildMemberNickname(v.GuildID, "@me", nickname)
}

// LeaveReporter
func LeaveReporter(v *voice.VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'leave'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	closeConnection(v)
	ChMessageSend(v.ChannelID, config.O.Greeting["leave"])
}

func closeConnection(v *voice.VoiceInstance) {
	time.Sleep(200 * time.Millisecond)
	v.Voice.Disconnect()
	log.Println("INFO: Voice channel destroyed")
	global.Mutex.Lock()
	delete(global.VoiceInstances, v.GuildID)
	global.Mutex.Unlock()
	Dg.UpdateGameStatus(0, config.O.Discord.Status)
	updateNickName(v, "")
}

func ListBotReporter(m *discordgo.MessageCreate) {
	botList, err := global.DB.ListBots(m.GuildID)
	if err != nil {
		return
	}

	msg := "```\n登録されているBOT一覧\n\n"
	for k, v := range botList {
		name := k
		botUser, err := Dg.User(k)
		if err == nil {
			name = botUser.Username
		} else {
			webhook, err := Dg.Webhook(k)
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
	botUser, err := Dg.User(commands[1])
	if err != nil {
		webHook, err := Dg.Webhook(commands[1])
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
	err = global.DB.AddBot(m.GuildID, commands[1], wavList)
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
	err := global.DB.DeleteBot(m.GuildID, commands[1])
	if err != nil {
		ChMessageSend(m.ChannelID, fmt.Sprintf("BOT ID「%s」の削除に失敗しました", commands[1]))
		return
	}
	ChMessageSend(m.ChannelID, fmt.Sprintf("BOT ID「%s」を削除しました", commands[1]))
}

func ListWordsReporter(m *discordgo.MessageCreate) {
	wordsList, err := global.DB.ListWords(m.GuildID)

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
	err := global.DB.AddWord(m.GuildID, commands[1], commands[2])
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
	err := global.DB.DeleteWord(m.GuildID, commands[1])
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
	user, err := Dg.User(userID)
	if err != nil {
		webHook, err := Dg.Webhook(userID)
		if err != nil {
			log.Println("ERROR: Cannot find user information.")
			return
		}
		user = webHook.User
	}
	DBUser, err := global.DB.GetUser(userID)
	userInfo := DBUser.UserInfo
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
		config.O.Discord.Prefix,
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
	user, _ := Dg.User(userID)
	if !user.Bot {
		log.Println("aaa")
		ChMessageSend(m.ChannelID, config.O.ErrorMsg["onlybot"])
		return
	}
	makeRandomHandlerInternal(userID, m)
}

func MakeRandomHandler(m *discordgo.MessageCreate) {
	makeRandomHandlerInternal(m.Author.ID, m)
}

func makeRandomHandlerInternal(userID string, m *discordgo.MessageCreate) {
	user := db.MakeRandom()
	global.DB.AddUser(userID, user)
	statusReporterInternal(userID, m)
}

func setStatusHandlerInternal(userID string, userInfo model.UserInfo, m *discordgo.MessageCreate) {
	_, ok := voice.Voices()[userInfo.Voice]
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

	if voice.IsVoiceRoid(userInfo.Voice) {
		if err := CheckRange(userInfo.Tone, 0.5, 2); err != nil {
			HelpReporter(m)
			return
		}
	}

	if err := CheckRange(userInfo.Intone, 0, 4); err != nil {
		HelpReporter(m)
		return
	}

	if voice.IsVoiceRoid(userInfo.Voice) {
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
	global.DB.AddUser(userID, userInfo)
	statusReporterInternal(userID, m)

}

func SetStatusForOtherHandler(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 9 {
		HelpReporter(m)
		return
	}
	userID := commands[1]

	user, err := Dg.User(userID)
	if err != nil {
		_, err := Dg.Webhook(userID)
		if err != nil {
			ChMessageSend(m.ChannelID, fmt.Sprintf("ID「%s」のBOTは見つかりませんでした。", userID))
			return
		}
	} else {
		if !user.Bot {
			ChMessageSend(m.ChannelID, config.O.ErrorMsg["onlyBot"])
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

	userInfo := model.UserInfo{}
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

	userInfo := model.UserInfo{}
	userInfo.Voice = voice
	userInfo.Speed, _ = strconv.ParseFloat(speed, 32)
	userInfo.Tone, _ = strconv.ParseFloat(tone, 32)
	userInfo.Intone, _ = strconv.ParseFloat(intone, 32)
	userInfo.Threshold, _ = strconv.ParseFloat(threshold, 32)
	userInfo.AllPass, _ = strconv.ParseFloat(allpass, 32)
	userInfo.Volume, _ = strconv.ParseFloat(volume, 32)

	setStatusHandlerInternal(m.Author.ID, userInfo, m)
}

func StopReporter(v *voice.VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'stop'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.Voice.ChannelID != voiceChannelID {
		return
	}
	v.StopTalking()
}

func RebootReporter(m *discordgo.MessageCreate) {
	commands := strings.Fields(m.Content)
	if len(commands) != 2 {
		return
	}
	secret := commands[1]
	if secret == config.O.Discord.Secret {
		panic("Rebooting")
	}
}

func SpeechText(v *voice.VoiceInstance, m *discordgo.MessageCreate) {
	content, err := m.Message.ContentWithMoreMentionsReplaced(v.Session)
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

	play.ReplaceWords(v.GuildID, &content)

	user, err := global.DB.GetUser(m.Author.ID)
	if err != nil {
		log.Println("INFO: Cannot Get User info")
		user, err = global.DB.NewUser(m.Author.ID)
		if err != nil {
			log.Println("ERR: Cannot initialize User")
			return
		}
	}
	botList, _ := global.DB.ListBots(v.GuildID)
	wavFileName := ""
	for k, v := range botList {
		if k == m.Author.ID {
			if len(v) != 0 {
				num := util.RandomInt(0, len(v))
				wavFileName = v[num]
			}
		}
	}
	speech := voice.Speech{content, user.UserInfo, wavFileName}
	speechSig := voice.SpeechSignal{speech, v}
	go func() {
		global.SpeechSignal <- speechSig
	}()
	// v.Talk(speech)
}

func CheckRange(val float64, min, max float64) error {
	if val < min || max < val {
		return errors.New("out of range")
	}
	return nil
}
