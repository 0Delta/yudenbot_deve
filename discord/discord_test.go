package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session
var c *discordgo.Channel

func Test_discodeInit(t *testing.T) {
	conf := GetToken("./.token.yml")
	t.Run("Create Instance", func(t *testing.T) {
		s = GetDiscord(conf.Token)
	})
	t.Run("Create Channel", func(t *testing.T) {
		c = CreateTextChannel(s, conf.GuildID, "go-test")
	})
	t.Run("Post", func(t *testing.T) {
		SendMessage(s, c, "テストメッセージです")
	})
}
