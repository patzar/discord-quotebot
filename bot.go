package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

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
}

type selector func([]string) (string, int)

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
			userQuote = UserQuotes{[]string{}, user}
		}
		userQuote.Quotes = append(userQuote.Quotes, text)
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
		if len(v) > 0 {
			var userQuote UserQuotes
			json.Unmarshal(v, &userQuote)
			if len(userQuote.Quotes) > 0 {
				q, i := fn(userQuote.Quotes)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> #%d: %s", i, q))
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
		i := len(quotes)-1
		return quotes[i], i 
	})
}
