package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
  "reflect"
  "strings"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
  bot Bot
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func parseCommand (s string) string {
  com := strings.Split(strings.TrimLeft(s, "."), " ")
  if len(com) > 0 {
	return strings.Title(com[0])
  }
  return ""
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
  inputs := []reflect.Value{reflect.ValueOf(s), reflect.ValueOf(m)}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
  method := reflect.ValueOf(bot).MethodByName(parseCommand(m.Content))
  if method.IsValid() && !method.IsZero() {
	  method.Call(inputs)
  }
}
