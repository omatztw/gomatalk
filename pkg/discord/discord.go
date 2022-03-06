package discord

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/omatztw/gomatalk/pkg/config"
	global "github.com/omatztw/gomatalk/pkg/global_vars"
	"github.com/omatztw/gomatalk/pkg/play"
	"github.com/omatztw/gomatalk/pkg/voice"
)

var (
	Dg *discordgo.Session
)

// DiscordConnect make a new connection to Discord
func DiscordConnect() (err error) {
	Dg, err = discordgo.New("Bot " + config.O.Discord.Token)
	if err != nil {
		log.Println("FATA: error creating Discord session,", err)
		return
	}
	log.Println("INFO: Bot is Opening")
	Dg.AddHandler(MessageCreateHandler)
	Dg.AddHandler(GuildCreateHandler)
	// dg.AddHandler(GuildDeleteHandler)
	Dg.AddHandler(VoiceStatusUpdateHandler)
	Dg.AddHandler(ConnectHandler)
	if config.O.Discord.NumShard > 1 {
		Dg.ShardCount = config.O.Discord.NumShard
		Dg.ShardID = config.O.Discord.ShardID
	}

	if config.O.Discord.Debug {
		Dg.LogLevel = discordgo.LogDebug
	}
	// Open Websocket
	err = Dg.Open()
	if err != nil {
		log.Println("FATA: Error Open():", err)
		return
	}
	_, err = Dg.User("@me")
	if err != nil {
		// Login unsuccessful
		log.Println("FATA:", err)
		return
	} // Login successful
	log.Println("INFO: Bot is now running. Press CTRL-C to exit.")
	initRoutine()
	Dg.UpdateGameStatus(0, config.O.Discord.Status)
	return nil
}

// SearchVoiceChannel search the voice channel id into from guild.
func SearchVoiceChannel(user string) (voiceChannelID string) {
	for _, g := range Dg.State.Guilds {
		for _, v := range g.VoiceStates {
			if v.UserID == user {
				return v.ChannelID
			}
		}
	}
	return ""
}

func UserCountVoiceChannel(voiceChannel string) int {
	count := 0
	for _, g := range Dg.State.Guilds {
		for _, v := range g.VoiceStates {
			user, _ := Dg.User(v.UserID)
			if !user.Bot {
				if v.ChannelID == voiceChannel {
					count++
				}
			}
		}
	}
	return count
}

// SearchGuild search the guild ID
func SearchGuild(textChannelID string) (guildID string) {
	channel, _ := Dg.Channel(textChannelID)
	guildID = channel.GuildID
	return
}

// ChMessageSend send a message and auto-remove it in a time
func ChMessageSend(textChannelID, message string) {
	for i := 0; i < 10; i++ {
		_, err := Dg.ChannelMessageSend(textChannelID, message)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}

func ChFileSend(textChannelID, name, message string) {
	Dg.ChannelFileSend(textChannelID, name, strings.NewReader(message))
}

// ChMessageSendEmbed send an embeded messages.
func ChMessageSendEmbed(textChannelID, title, description string, user discordgo.User) {
	embed := discordgo.MessageEmbed{}
	embed.Title = title
	embed.Description = description
	embed.Color = 0xb20000
	author := discordgo.MessageEmbedAuthor{}
	author.Name = user.Username
	author.IconURL = user.AvatarURL("")
	embed.Author = &author
	for i := 0; i < 10; i++ {
		_, err := Dg.ChannelMessageSendEmbed(textChannelID, &embed)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}

func initRoutine() {
	global.SpeechSignal = make(chan voice.SpeechSignal)
	go play.GlobalPlay(global.SpeechSignal)
}

// ConnectHandler
func ConnectHandler(s *discordgo.Session, connect *discordgo.Connect) {
	s.UpdateGameStatus(0, config.O.Discord.Status)
}

// GuildCreateHandler
func GuildCreateHandler(s *discordgo.Session, guild *discordgo.GuildCreate) {
	log.Println("INFO: Guild Create:", guild.ID)
	err := global.DB.CreateGuild(guild.ID)
	if err != nil {
		log.Println("FATA: DB", err)
		return
	}
}

func VoiceStatusUpdateHandler(s *discordgo.Session, voice *discordgo.VoiceStateUpdate) {
	v := global.VoiceInstances[voice.GuildID]
	if v == nil {
		return
	}
	if v.Voice == nil {
		return
	}
	user, _ := Dg.User(voice.UserID)
	botUser, _ := Dg.User("@me")

	if voice.UserID == botUser.ID {
		if voice == nil || voice.BeforeUpdate == nil || voice.ChannelID == "" {
			return
		}
		if voice.BeforeUpdate.ChannelID != voice.ChannelID {
			v.Voice, _ = Dg.ChannelVoiceJoin(v.GuildID, voice.ChannelID, false, false)
		}
	}

	if user.Bot && voice.UserID != botUser.ID {
		// Ignore Bot
		return
	}

	userCount := UserCountVoiceChannel(v.Voice.ChannelID)
	if userCount == 0 {
		v.Lock()
		defer v.Unlock()
		if v.Voice == nil {
			log.Println("INFO: Voice channel has already been destroyed")
			return
		}
		if v.Session.VoiceConnections[v.GuildID] != nil {
			v.Voice.Disconnect()
			log.Println("INFO: Voice channel destroyed")
			global.Mutex.Lock()
			delete(global.VoiceInstances, v.GuildID)
			global.Mutex.Unlock()
			updateNickName(v, "")
			ChMessageSend(v.ChannelID, config.O.Greeting["nobody"])
		}
	}
}

// MessageCreateHandler
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := SearchGuild(m.ChannelID)
	botList, _ := global.DB.ListBots(guildID)
	isSpecial := false
	if m.Author.Bot {
		if _, ok := botList[m.Author.ID]; !ok {
			return
		}
		isSpecial = true
	}
	v := global.VoiceInstances[guildID]
	if strings.HasPrefix(m.Content, config.O.Discord.Prefix) {
		content := strings.Replace(m.Content, config.O.Discord.Prefix, "", 1)
		command := strings.Fields(content)

		if len(command) == 0 {
			return
		}

		switch command[0] {
		case "help", "h":
			HelpReporter(m)
		case "summon", "s":
			JoinReporter(v, m, s)
		case "bye", "b":
			LeaveReporter(v, m)
		case "stop":
			StopReporter(v, m)
		case "words_list", "wl":
			ListWordsReporter(m)
		case "add_word", "aw":
			AddWordReporter(m)
		case "delete_word", "dw":
			DeleteWordReporter(m)
		case "status":
			StatusReporter(m)
		case "update_voice", "uv":
			SetStatusHandler(m)
		case "add_bot", "ab":
			AddBotReporter(m)
		case "delete_bot", "db":
			DeleteBotReporter(m)
		case "bots_list", "bl":
			ListBotReporter(m)
		case "random", "r":
			MakeRandomHandler(m)
		case "update_bot_voice", "ubv":
			SetStatusForOtherHandler(m)
		case "random_bot", "rb":
			MakeRandomForOther(m)
		case "reboot":
			RebootReporter(m)
		default:
			return
		}
		return
	}
	if v != nil && v.Voice != nil {
		if !isSpecial && v.ChannelID != m.ChannelID {
			return
		}
		SpeechText(v, m)
	}
}
