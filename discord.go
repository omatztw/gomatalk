package main

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// DiscordConnect make a new connection to Discord
func DiscordConnect() (err error) {
	dg, err = discordgo.New("Bot " + o.DiscordToken)
	if err != nil {
		log.Println("FATA: error creating Discord session,", err)
		return
	}
	log.Println("INFO: Bot is Opening")
	dg.AddHandler(MessageCreateHandler)
	dg.AddHandler(GuildCreateHandler)
	// dg.AddHandler(GuildDeleteHandler)
	dg.AddHandler(VoiceStatusUpdateHandler)
	dg.AddHandler(ConnectHandler)
	// Open Websocket
	err = dg.Open()
	if err != nil {
		log.Println("FATA: Error Open():", err)
		return
	}
	_, err = dg.User("@me")
	if err != nil {
		// Login unsuccessful
		log.Println("FATA:", err)
		return
	} // Login successful
	log.Println("INFO: Bot user test")
	log.Println("INFO: Bot is now running. Press CTRL-C to exit.")
	initRoutine()
	dg.UpdateStatus(0, o.DiscordStatus)
	return nil
}

// SearchVoiceChannel search the voice channel id into from guild.
func SearchVoiceChannel(user string) (voiceChannelID string) {
	for _, g := range dg.State.Guilds {
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
	for _, g := range dg.State.Guilds {
		for _, v := range g.VoiceStates {
			user, _ := dg.User(v.UserID)
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
	channel, _ := dg.Channel(textChannelID)
	guildID = channel.GuildID
	return
}

// ChMessageSend send a message and auto-remove it in a time
func ChMessageSend(textChannelID, message string) {
	for i := 0; i < 10; i++ {
		_, err := dg.ChannelMessageSend(textChannelID, message)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
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
		_, err := dg.ChannelMessageSendEmbed(textChannelID, &embed)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}

func initRoutine() {
	speechSignal = make(chan SpeechSignal)
	go GlobalPlay(speechSignal)
}

// ConnectHandler
func ConnectHandler(s *discordgo.Session, connect *discordgo.Connect) {
	log.Println("INFO: Connected!!")
	s.UpdateStatus(0, o.DiscordStatus)
}

// GuildCreateHandler
func GuildCreateHandler(s *discordgo.Session, guild *discordgo.GuildCreate) {
	log.Println("INFO: Guild Create:", guild.ID)
	err := CreateGuildDB(guild.ID)
	if err != nil {
		log.Println("FATA: DB", err)
		return
	}
}

func VoiceStatusUpdateHandler(s *discordgo.Session, voice *discordgo.VoiceStateUpdate) {
	v := voiceInstances[voice.GuildID]
	if v == nil {
		return
	}
	if v.voice == nil {
		return
	}
	user, _ := dg.User(voice.UserID)
	if user.Bot {
		// Ignore Bot
		return
	}
	userCount := UserCountVoiceChannel(v.voice.ChannelID)
	log.Println("Status Update", userCount)
	if userCount == 0 {
		v.voice.Disconnect()
		log.Println("INFO: Voice channel destroyed")
		mutex.Lock()
		delete(voiceInstances, v.guildID)
		mutex.Unlock()
		dg.UpdateStatus(0, o.DiscordStatus)
		ChMessageSend(v.channelID, "すやぁ")
	}
}

// // GuildDeleteHandler
// func GuildDeleteHandler(s *discordgo.Session, guild *discordgo.GuildDelete) {
// 	log.Println("INFO: Guild Delete:", guild.ID)
// 	v := voiceInstances[guild.ID]
// 	if v != nil {
// 		v.Stop()
// 		time.Sleep(200 * time.Millisecond)
// 		mutex.Lock()
// 		delete(voiceInstances, guild.ID)
// 		mutex.Unlock()
// 	}
// }

// MessageCreateHandler
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.Bot {
		return
	}

	guildID := SearchGuild(m.ChannelID)
	v := voiceInstances[guildID]
	if strings.HasPrefix(m.Content, o.DiscordPrefix) {
		content := strings.Replace(m.Content, o.DiscordPrefix, "", 1)
		command := strings.Fields(content)

		if len(command) == 0 {
			return
		}

		switch command[0] {
		case "help", "h":
			HelpReporter(m)
		case "summon", "s":
			JoinReporter(v, m, s)
		case "leave", "l":
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
		default:
			return
		}
		return
	}
	if v != nil {
		SpeechText(v, m)
	}
}
