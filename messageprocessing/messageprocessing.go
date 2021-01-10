package messageprocessing

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/poodlenoodle42/Discord_Calender_Bot/database"
)

type apWithChannel struct {
	ap database.Appointment
	ch database.Channel
}

//Lookup table to all the Channels a person is memeber in, to recieve appointment information from database
var Lookup map[string][]database.Channel

//KnownChannels List of all ChannelIDs the bot has appointments of
var KnownChannels map[string]struct{}

//ValidTypes for appointment, to be populated from config
var ValidTypes []string

//InitLookups inits the Lookups with the data from the database
func InitLookups() {
	KnownChannels, Lookup = database.PopulateLookups()
}

func apointmentToString(ap apWithChannel) string {
	return fmt.Sprintf(`%s: \n
	\t %s \n
	\t %s \n
	\t %s \n`, ap.ch.ID, ap.ap.Deadline.Format("02.01.2006 15:04"), ap.ap.Ty, ap.ap.Description)
}

//GetAppointments is called when a private message is recieved, all apointments for the author are send
func GetAppointments(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, ex := Lookup[m.Author.ID]
	if !ex {
		log.Printf("Unknown User with ID %s on Channel with ID %s \n", m.Author.ID, m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "An error occourd, you are not known to the bot")
		return
	}
	aps := make([]apWithChannel, 0, 50)
	for _, ch := range Lookup[m.Author.ID] {
		apsN, err := database.GetAppointmentsFromDatabase(ch.ID)
		apsWC := make([]apWithChannel, 0, len(apsN))
		if err != nil {
			log.Printf("Error getting appointments from db, %s \n", err.Error())
			s.ChannelMessageSend(m.ChannelID, "An error occourd, some appointments could not be recieved from the db")
			return
		}
		for _, ap := range apsN {
			apsWC = append(apsWC, apWithChannel{ap: ap, ch: ch})
		}
		aps = append(aps, apsWC...)
	}
	sort.Slice(aps, func(a int, b int) bool {
		return aps[a].ap.Deadline.Before(aps[b].ap.Deadline)
	})
	var toSend string
	for _, ap := range aps {
		if ap.ap.Deadline.After(time.Now()) {
			toSend += apointmentToString(ap)
		}
	}
	s.ChannelMessageSend(m.ChannelID, toSend)
}

func isTypeValid(ty string) bool {
	for _, t := range ValidTypes {
		if t == ty {
			return true
		}
	}
	return false
}

//SetAppointment is called when a "public" message is recieved attempting to set a new apointment
func SetAppointment(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, ex := KnownChannels[m.ChannelID]
	if !ex {
		newChannel(s, m)
	}
	message := strings.Split(m.Content[1:], " ")
	description := strings.Join(message[4:], " ")
	timeV := strings.Join(message[2:3], " ")
	ty := message[1]
	action := message[0]
	var ap database.Appointment
	if !isTypeValid(ty) {
		s.ChannelMessageSend(m.ChannelID, "Invalid Type")
		return
	}
	t, err := time.Parse("02.01.2006 15:04", timeV)
	if err != nil {
		log.Printf("Error parsing time %s, Error %s\n", timeV, err.Error())
		s.ChannelMessageSend(m.ChannelID, "Time could not be parsed")
		return
	}
	ap.Deadline = t
	ap.Description = description
	ap.Ty = ty
	if action == "delete" {
		err = database.DeleteAppointment(m.ChannelID, ap)
		if err != nil {
			log.Printf("Error deleting appointment, %s \n", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Could not delete appointment")
			return
		}
	} else if action == "add" {
		err = database.WriteAppointmentToDatabse(m.ChannelID, ap)
		if err != nil {
			log.Printf("Error writing appointment, %s \n", err.Error())
			s.ChannelMessageSend(m.ChannelID, "Could not write appointment")
			return
		}
	} else {
		log.Printf("Wrong action %s \n", action)
		s.ChannelMessageSend(m.ChannelID, "Action not known")
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
