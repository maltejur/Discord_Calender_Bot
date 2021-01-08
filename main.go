package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/poodlenoodle42/Discord_Calender_Bot/config"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.GuildID == "" {
		fmt.Println("Private Message")
	}
}

func main() {
	config := config.ReadConfigFile("config.yaml")
	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(err)
	}
	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		panic(err)
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
