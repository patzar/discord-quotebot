package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

// Bot holds the bot state.
type Bot struct {
	db *bolt.DB
}

// UserQuotes stores quotes of a particular user.
type UserQuotes struct {
	Quotes   []string
	Username string
	QuoteIDs map[string]string
}

type selector func([]string) (string, int)

func first(arr []*discordgo.MessageEmbed) *discordgo.MessageEmbed {
	if len(arr) > 0 {
		return arr[0]
	}
	return nil
}

// Saves quote to the database. Returns true if the save was successful.
func (b Bot) quoteImpl(user string, text string, messageID string) bool {
	saved := false

	// Check whether particular message was already saved to the database.
	b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("QuoteSavedBucket"))
		v := bucket.Get([]byte(messageID))
		if len(v) > 0 {
			saved = true
			return err
		}
		err = bucket.Put([]byte(messageID), []byte{1})
		return err
	})

	if saved {
		return false
	}

	// Save the message to the database.
	b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("QuoteBucket"))
		v := bucket.Get([]byte(user))
		var userQuote UserQuotes
		if len(v) > 0 {
			json.Unmarshal(v, &userQuote)
		} else {
			userQuote = UserQuotes{[]string{}, user, map[string]string{}}
		}
		userQuote.Quotes = append(userQuote.Quotes, text)
		userQuote.QuoteIDs[text] = messageID
		userQuoteSerialized, err := json.Marshal(userQuote)

		err = bucket.Put([]byte(user), []byte(userQuoteSerialized))
		return err
	})
	return true
}

// Views a quote for a given user, using selector of which quote in the array to show.
func (b Bot) viewQuote(s *discordgo.Session, m *discordgo.MessageCreate, user string, fn selector) {
	b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("QuoteBucket"))
		v := bucket.Get([]byte(user))
		sent := false
		if len(v) > 0 {
			var userQuote UserQuotes
			json.Unmarshal(v, &userQuote)
			if len(userQuote.Quotes) > 0 {
				q, i := fn(userQuote.Quotes)
				qid, ok := userQuote.QuoteIDs[q]
				fmt.Println(qid)
				if ok {
					msg, err := s.ChannelMessage(m.ChannelID, qid)
					if err == nil {
						a := msg.Attachments[0]
						resp, _ := http.Get(a.URL)
						//defer resp.Body.Close()
						files := []*discordgo.File{}
						if err == nil {
							files = append(files, &discordgo.File{
								Name:        "blah.png",
								ContentType: "image/png",
								Reader:      resp.Body,
							})
						}
						s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
							Content: fmt.Sprintf("> #%d: %s", i, q),
							Embed:   first(msg.Embeds),
							Files:   files,
						})
						sent = true

					}
				}
				if !sent {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> #%d: %s", i, q))
				}
			}
		}
		return nil
	})
}

// Randomquote views a random quote.
func (b Bot) Randomquote(s *discordgo.Session, m *discordgo.MessageCreate, user string) {
	b.viewQuote(s, m, user, func(quotes []string) (string, int) {
		i := rand.Intn(len(quotes))
		return quotes[i], i
	})
}

// Lastquote views a last quote.
func (b Bot) Lastquote(s *discordgo.Session, m *discordgo.MessageCreate, user string) {
	b.viewQuote(s, m, user, func(quotes []string) (string, int) {
		i := len(quotes) - 1
		return quotes[i], i
	})
}
