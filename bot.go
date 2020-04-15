package main
import (
	"github.com/bwmarrin/discordgo"
)

type Bot struct {}

func (b Bot) Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
    s.ChannelMessageSend(m.ChannelID, "Pong!")
}