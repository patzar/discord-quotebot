package main
import (
	"github.com/bwmarrin/discordgo"
)

type Bot struct {}

func (b Bot) Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
    s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func (b Bot) Text(s *discordgo.Session, m *discordgo.MessageCreate, t string) {
    s.ChannelMessageSend(m.ChannelID, t)
}