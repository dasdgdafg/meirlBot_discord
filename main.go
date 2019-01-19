package main

// based on https://github.com/bwmarrin/discordgo/blob/master/examples/pingpong/main.go

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var passwordBytes, _ = ioutil.ReadFile("password.txt")
var password = string(passwordBytes)

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + password)
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

var cuteImage = CuteImage{}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	if cuteImage.checkForMatch(m.Content) {
		go sendImage(s, m.ChannelID, m.Content, m.Author.Username, cuteImage)
	}
}

func sendImage(s *discordgo.Session, sendTo string, msg string, nick string, img CuteImage) {
	str, url, err := img.getImageForMessage(msg, nick)
	var newMsg string
	if err != nil {
		newMsg = "error fetching image"
	} else if url == "" {
		newMsg = "couldn't find any images"
	} else {
		newMsg = str + " " + url
	}
	log.Println("sending to: " + sendTo + ", message: " + newMsg)
	s.ChannelMessageSend(sendTo, newMsg)
}
