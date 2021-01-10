package messageprocessing

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/poodlenoodle42/Discord_Calender_Bot/database"
)

//Lookup table to all the Channels a person is memeber in, to recieve appointment information from database
var Lookup map[string][]database.Channel

//KnownChannels List of all ChannelIDs the bot has appointments of
var KnownChannels map[string]struct{}

//InitLookups inits the Lookups with the data from the database
func InitLookups() {
	KnownChannels, Lookup = database.PopulateLookups()
}

//GetAppointments is called when a private message is recieved, all apointments for the author are send
func GetAppointments(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, ex := Lookup[m.Author.ID]
	if !ex {
		log.Printf("Unknown User with ID %s on Channel with ID %s \n", m.Author.ID, m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "An error occourd, you are not known to the bot")
		return
	}

}

//SetAppointment is called when a "public" message is recieved attempting to set a new apointment
func SetAppointment(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, ex := KnownChannels[m.ChannelID]
	if !ex {
		newChannel(s, m)
	}

}

func newChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	//New Channel
	log.Printf("New Channel with ID %s\n", m.ChannelID)
	c, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Printf("Error recieving Channel Information, Error: %s \n", err.Error())
		s.ChannelMessageSend(m.ChannelID, "An error occourd, the bot could not recieve Channel info")
		return
	}
	err = database.MakeNewChannelTable(c.ID)
	if err != nil {
		log.Printf("Error making table in database: %s \n", err.Error())
		s.ChannelMessageSend(m.ChannelID, "An error occourd, the bot could not make channel table")
		return
	}
	database.WriteNewID(c.ID, c.Name, database.ChannelC)
	ch := database.Channel{ID: c.ID, Name: c.Name}
	KnownChannels[c.ID] = struct{}{}
	er := false
	for _, user := range c.Recipients {
		_, ex := Lookup[user.ID]
		if !ex {
			err = database.WriteNewID(user.ID, user.Username, database.UserC)
			if err != nil {
				log.Printf("Error writing new userID in database: %s \n", err.Error())
				er = true
			}
			err = database.NewUserTable(user.ID)
			if err != nil {
				log.Printf("Error making User table in database: %s \n", err.Error())
				er = true
			}
		}
		Lookup[user.ID] = append(Lookup[user.ID], ch)
		err = database.WriteNewLookupEntry(user.ID, ch)
		if err != nil {
			log.Printf("Error writing new lookupentry in database: %s \n", err.Error())
			er = true
		}
	}
	if er {
		s.ChannelMessageSend(m.ChannelID, "An error occourd, adding new users to the lookupdb")
	}
}
