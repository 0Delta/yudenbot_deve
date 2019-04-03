package discord

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

var session *discordgo.Session
var m sync.Mutex

type DiscordAuth struct {
	Token   string `yaml:"discordtoken" json:"discordtoken"`
	GuildID string `yaml:"guildid" json:"guildid"`
}

func GetDiscord(token string) *discordgo.Session {
	m.Lock()
	defer m.Unlock()
	if session == nil {
		discodeInit(token)
	}
	return session
}

func GetToken(filepath string) *DiscordAuth {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("error: Error while load token : ", err)
	}
	var auth DiscordAuth
	err = yaml.Unmarshal(buf, &auth)
	if err != nil {
		log.Fatal("error: Error while unmarshal token: ", err)
	}
	return &auth
}

func discodeInit(token string) {

	Token := "Bot " + token
	discord, err := discordgo.New()
	if err != nil {
		log.Fatalln("error: Error logging in : ", err)
	}
	discord.Token = Token

	// discord.AddHandler(onMessageCreate) //全てのWSAPIイベントが発生した時のイベントハンドラを追加
	// websocketを開いてlistening開始
	err = discord.Open()
	if err != nil {
		log.Fatal("error: Error open discord", err)
	}
	session = discord
}

//メッセージを送信する関数
func SendMessage(s *discordgo.Session, c *discordgo.Channel, msg string) *discordgo.Message {
	newMes, err := s.ChannelMessageSend(c.ID, msg)

	if err != nil {
		log.Println("error: Error sending message: ", err)
		return nil
	}
	log.Println("info: Post Message : " + msg)
	return newMes
}

func CreateTextChannel(s *discordgo.Session, guildid string, name string) *discordgo.Channel {
	newCh, err := s.GuildChannelCreate(guildid, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Println("error: Error Create channel : ", err)
		return nil
	}
	log.Println("info: Create Channel : " + name + "@" + newCh.ID)
	return newCh
}
