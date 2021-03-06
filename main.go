package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strings"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/logger"
)

// Variables used for command line parameters
var (
	Token string
	bot   Bot
)

const logPath = "bot.log"

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	fmt.Println(*lf)
	defer lf.Close()
	defer logger.Init("LoggerExample", false, false, lf).Close()

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	bot = Bot{db}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		logger.Fatalf("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageReactionAdd)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		logger.Fatalf("error opening connection,", err)
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

// Converts input string to a reflection-compatible corresponding method string, e.g.
// .foo bar -> Foo
func parseCommand(s string) string {
	com := strings.Split(strings.TrimLeft(s, "."), " ")
	if len(com) > 0 {
		return strings.Title(com[0])
	}
	return ""
}

// Returns all command arguments, i.e. all words except from the first one.
// If the arguments include brackets, consider them as one arguments, e.g.
// .foo bar "hello world" -> ["bar", "hello world"]
func parseArguments(s string) []string {
	re := regexp.MustCompile(`[^\s"']+|([^\s"']*"([^"]*)"[^\s"']*)+|'([^']*)`)
	args := re.FindAllString(s, -1)
	if len(args) == 0 {
		return []string{}
	}
	return args[1:]
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	fmt.Println(m.Emoji.Name)
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	// Ignore all messages created by the bot itself
	if err != nil || message.Author.ID == s.State.User.ID {
		return
	}
	if m.Emoji.Name == "❤️" {
		if err != nil {
			logger.Fatalf("error loading Discord message,", err)
			return
		}
		saved := bot.quoteImpl(message.Author.Username, message.Content, m.MessageID)
		if saved {
			quote := fmt.Sprintf("Saved quote: \n > %s \n > -%s", message.Content, message.Author.Username)
			s.ChannelMessageSend(m.ChannelID, quote)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger.Info(m.Content)
	args := parseArguments(m.Content)
	// Pass the original session and message arguments.
	inputs := make([]reflect.Value, len(args)+2)
	inputs[0] = reflect.ValueOf(s)
	inputs[1] = reflect.ValueOf(m)
	// Pass any additional arguments based on the message itself.
	for i, _ := range args {
		inputs[i+2] = reflect.ValueOf(args[i])
	}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	method := reflect.ValueOf(bot).MethodByName(parseCommand(m.Content))
	if method.IsValid() && len(inputs) >= method.Type().NumIn() {
		// Trim all unnecessary arguments.
		inputs = inputs[:method.Type().NumIn()]
		method.Call(inputs)
	}
}
