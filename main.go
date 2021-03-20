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

var configVar config.Config

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
	if m.GuildID == configVar.AdminChannelID && m.Content[1] != '-' {
		if m.Content[2:] == "reg" {
			regenerateLookup(s)
		}
	}
	if m.GuildID == "" { //Private message
		if m.Content[1:] == "help" {
			s.ChannelMessageSend(m.ChannelID, configVar.HelpMessage)
		} else {
			messageprocessing.GetAppointments(s, m)
		}
	} else {
		//$Action type dd.mm.yyyy hh:mm description
		var botM = messageprocessing.SetAppointment(s, m)
		if configVar.DeleteMessages {
			time.AfterFunc(time.Duration(configVar.WaitBeforeDelete)*time.Second, func() {
				err := s.ChannelMessageDelete(m.ChannelID, m.ID)
				if err != nil {
					log.Printf("Message could not be deleted: %s\n", err.Error())
				}
				if botM != nil {
					err = s.ChannelMessageDelete(botM.ChannelID, botM.ID)
					if err != nil {
						log.Printf("Message could not be deleted: %s\n", err.Error())
					}
				}
			})
		}
	}

}

func regenerateLookup(s *discordgo.Session) {
	log.Printf("Start regenerateLookup")
	messageprocessing.KnownChannels = make(map[string]struct{})
	messageprocessing.Lookup = make(map[string][]database.Channel)
	err := database.ClearLookupDB()
	gs, err := s.UserGuilds(100, "", "")
	if err != nil {
		log.Printf("Error geting Guilds of Bot: %s\n", err.Error())
	}
	for _, g := range gs {
		messageprocessing.PopulateLookupForGuild(g.ID, s)
	}
	log.Printf("End regenerateLookup")
}

func main() {
	messageprocessing.Lookup = make(map[string][]database.Channel)
	messageprocessing.KnownChannels = make(map[string]struct{})
	config := config.ReadConfigFile("config.yaml")
	//OpenLogFile
	f, err := os.OpenFile(config.Logfile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	messageprocessing.ValidTypes = config.Validtypes
	configVar = config
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
	go func() {
		for {
			time.Sleep(time.Duration(config.RegenerateLookupTime) * time.Hour)
			regenerateLookup(discord)
		}
	}()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
