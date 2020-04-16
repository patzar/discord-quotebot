package main
import (
	"github.com/bwmarrin/discordgo"
  "github.com/boltdb/bolt"
  "fmt"
  "encoding/json"
  "math/rand"
)

type Bot struct {
  db *bolt.DB
}

type UserQuotes struct {
  Quotes []string
  Username string
}

type selector func([]string) string

func (b Bot) quoteImpl(user string, text string) {
    b.db.Update(func(tx *bolt.Tx) error {
    bucket, err := tx.CreateBucketIfNotExists([]byte("QuoteBucket"))
    v := bucket.Get([]byte(user))
    var userQuote UserQuotes
    if len(v) > 0 {
	    json.Unmarshal(v, &userQuote)
    } else {
      userQuote = UserQuotes{[]string{}, user}
    }
    userQuote.Quotes = append(userQuote.Quotes, text)
    userQuoteSerialized, err := json.Marshal(userQuote)

    err = bucket.Put([]byte(user), []byte(userQuoteSerialized))
    return err
  })
}

func (b Bot) viewQuote(s *discordgo.Session, m *discordgo.MessageCreate, user string, fn selector) {
  b.db.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("QuoteBucket"))
    v := bucket.Get([]byte(user))
    if len(v) > 0 {
      var userQuote UserQuotes
	    json.Unmarshal(v, &userQuote)
      if len(userQuote.Quotes) > 0 {
        q := fn(userQuote.Quotes)
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> %s", q))
      }
    }
    return nil
  })
}

func (b Bot) Randomquote(s *discordgo.Session, m *discordgo.MessageCreate, user string) {
  b.viewQuote(s, m, user, func(quotes []string) string {
    return quotes[rand.Intn(len(quotes))]
  })
}

func (b Bot) Lastquote(s *discordgo.Session, m *discordgo.MessageCreate, user string) {
    b.viewQuote(s, m, user, func(quotes []string) string {
    return quotes[len(quotes)-1]
  })
}