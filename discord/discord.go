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
		log.Fatal("Error while load token : ", err)
	}
	var auth DiscordAuth
	err = yaml.Unmarshal(buf, &auth)
	if err != nil {
		log.Fatal("Error while unmarshal token: ", err)
	}
	return &auth
}

func discodeInit(token string) {

	Token := "Bot " + token
	discord, err := discordgo.New()
	if err != nil {
		log.Println("Error logging in")
		log.Fatal(err)
	}
	discord.Token = Token

	// discord.AddHandler(onMessageCreate) //全てのWSAPIイベントが発生した時のイベントハンドラを追加
	// websocketを開いてlistening開始
	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}
	session = discord
}

//メッセージを送信する関数
func SendMessage(s *discordgo.Session, c *discordgo.Channel, msg string) *discordgo.Message {
	newMes, err := s.ChannelMessageSend(c.ID, msg)

	log.Println(">>> " + msg)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
	return newMes
}

func CreateTextChannel(s *discordgo.Session, guildid string, name string) *discordgo.Channel {
	newCh, err := s.GuildChannelCreate(guildid, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Println("Error Create channel : ", err)
		return nil
	}
	log.Println("Create Channel : " + name + "@" + newCh.ID)
	return newCh
}
