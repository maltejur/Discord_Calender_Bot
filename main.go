package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/poodlenoodle42/Discord_Calender_Bot/config"
	"github.com/poodlenoodle42/Discord_Calender_Bot/database"
	"github.com/poodlenoodle42/Discord_Calender_Bot/messageprocessing"
)

var helpMessage string

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "" {
		return
	}
	if m.Content[0] != '$' {
		return
	}
	if m.GuildID == "" { //Private message
		if m.Content[1:] == "help" {
			s.ChannelMessageSend(m.ChannelID, helpMessage)
		} else {
			messageprocessing.GetAppointments(s, m)
		}
	} else {
		//$Action type dd.mm.yyyy hh:mm description
		var botM = messageprocessing.SetAppointment(s, m)
		time.AfterFunc(3*time.Second, func() {
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				log.Printf("Message could not be deleted: %s", err.Error())
			}
			if botM != nil {
				s.ChannelMessageDelete(botM.ChannelID, botM.ID)
				if err != nil {
					log.Printf("Message could not be deleted: %s", err.Error())
				}
			}
		})
	}

}

func main() {
	messageprocessing.Lookup = make(map[string][]database.Channel)
	messageprocessing.KnownChannels = make(map[string]struct{})
	config := config.ReadConfigFile("config.yaml")
	//OpenLogFile
	f, err := os.Create(config.Logfile)
	if err != nil {
		panic(err)
	}
	messageprocessing.ValidTypes = config.Validtypes
	helpMessage = config.HelpMessage
	log.SetOutput(f)
	log.Println("Start")
	defer log.Println("End")
	defer f.Close()
	database.InitDB(config.DatabaseFile, config.LookupDatabaseFile)
	messageprocessing.InitLookups()

	discord, err := discordgo.New("Bot " + config.Token)
	//discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
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
