package main
import (
	"github.com/bwmarrin/discordgo"
  "github.com/boltdb/bolt"
  "fmt"
)

type Bot struct {
  db *bolt.DB
}

func (b Bot) Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
    s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func (b Bot) Text(s *discordgo.Session, m *discordgo.MessageCreate, t string) {
    s.ChannelMessageSend(m.ChannelID, t)
}

func (b Bot) Quote(s *discordgo.Session, m *discordgo.MessageCreate, user string, text string) {
    b.db.Update(func(tx *bolt.Tx) error {
    bucket, err := tx.CreateBucketIfNotExists([]byte("QuoteBucket"))
    err = bucket.Put([]byte(user), []byte(text))
    return err
  })
    s.ChannelMessageSend(m.ChannelID, "Saved quote.")
}

func (b Bot) Retrieve(s *discordgo.Session, m *discordgo.MessageCreate, user string) {
  b.db.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("QuoteBucket"))
    v := bucket.Get([]byte(user))
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s", v))
    return nil
  })
}