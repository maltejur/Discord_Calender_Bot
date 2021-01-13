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
	return fmt.Sprintf(`%s: 
	%s 
	%s 
	%s
`, ap.ch.Name, ap.ap.Deadline.Format("02.01.2006 15:04"), ap.ap.Ty, ap.ap.Description)
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
		ch, _ := s.Channel(m.ChannelID)
		populateLookupForGuild(ch, s)
	}
	message := strings.Split(m.Content[1:], " ")
	description := strings.Join(message[4:], " ")
	timeV := strings.Join(message[2:4], " ")
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
		return
	}
}

func newUserForChannel(user *discordgo.User, channel *discordgo.Channel) error {
	_, ex := Lookup[user.ID]
	if !ex {
		err := database.WriteNewID(user.ID, user.Username, database.UserC)
		if err != nil {
			log.Printf("Error writing new userID in database: %s \n", err.Error())
			return err
		}
		err = database.NewUserTable(user.ID)
		if err != nil {
			log.Printf("Error making User table in database: %s \n", err.Error())
			return err
		}
	}
	ch := database.Channel{ID: channel.ID, Name: channel.Name}
	Lookup[user.ID] = append(Lookup[user.ID], ch)
	err := database.WriteNewLookupEntry(user.ID, ch)
	if err != nil {
		log.Printf("Error writing NewLookupEntry: %s \n", err.Error())
	}
	return err
}

func populateLookupForGuild(c *discordgo.Channel, s *discordgo.Session) {
	members, err := s.GuildMembers(c.GuildID, "", 1000)
	if err != nil {
		log.Printf("Error getting members: %s \n", err.Error())
		s.ChannelMessageSend(c.ID, "An error occourd, could not get members of guild")
		return
	}
	channels, err := s.GuildChannels(c.GuildID)
	if err != nil {
		log.Printf("Error getting channels: %s \n", err.Error())
		s.ChannelMessageSend(c.ID, "An error occourd, the bot could not get the channels of the guild")
		return
	}
	for _, channel := range channels {
		_, ex := KnownChannels[channel.ID]
		if !ex {
			err := database.MakeNewChannelTable(channel.ID)
			if err != nil {
				log.Printf("Error making table in database: %s \n", err.Error())
				s.ChannelMessageSend(c.ID, "An error occourd, the bot could not make channel table")
				return
			}
			err = database.WriteNewID(channel.ID, channel.Name, database.ChannelC)
			if err != nil {
				log.Printf("Error writing new Channel: %s \n", err.Error())
				s.ChannelMessageSend(c.ID, "An error occourd, could not write new ChannelID")
				return
			}
			KnownChannels[channel.ID] = struct{}{}
			for _, member := range members {
			PermissionBreak:
				for _, permission := range channel.PermissionOverwrites {
					if permission.Type == "role" { //Role
						for _, roleID := range member.Roles {
							if roleID == permission.ID && permission.Allow&0x00000400 == 0x00000400 {
								err = newUserForChannel(member.User, channel)
								if err != nil {
									log.Printf("Error writing new User for Channel: %s \n", err.Error())
									s.ChannelMessageSend(c.ID, "An error occourd, could not write new User for Channel")
									return
								}
								break PermissionBreak
							}
						}
					} else if permission.Type == "member" { //Per Person
						if member.User.ID == permission.ID && permission.Allow&0x00000400 == 0x00000400 {
							err = newUserForChannel(member.User, channel)
							if err != nil {
								log.Printf("Error writing new User for Channel: %s \n", err.Error())
								s.ChannelMessageSend(c.ID, "An error occourd, could not write new User for Channel")
								return
							}
							break PermissionBreak
						}
					}
				}
			}
		}
	}

}
