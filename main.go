package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/poodlenoodle42/Discord_Calender_Bot/config"
	"github.com/poodlenoodle42/Discord_Calender_Bot/messageprocessing"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.GuildID == "" { //Private message
		messageprocessing.GetAppointments(s, m)
	} else {
		messageprocessing.SetAppointment(s, m)
	}

}

func main() {
	messageprocessing.Lookup = make(map[string][]messageprocessing.Channel)
	config := config.ReadConfigFile("config.yaml")

	//OpenLogFile
	f, err := os.Create(config.Logfile)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	log.Println("Start")
	defer log.Println("End")
	defer f.Close()

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Panic(err)
	}
	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
