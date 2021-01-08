package messageprocessing

import "github.com/bwmarrin/discordgo"

//Channel Simplification of discordgo.Channel
type Channel struct {
	ID   string
	name string
}

//Lookup table to all the Channels a person is memeber in, to recieve appointment information from database
var Lookup map[string]map[string][]Channel

//GetAppointments is called when a private message is recieved, all apointments for the author are send
func GetAppointments(s *discordgo.Session, m *discordgo.MessageCreate) {

}

//SetAppointment is called when a "public" message is recieved attempting to set a new apointment
func SetAppointment(s *discordgo.Session, m *discordgo.MessageCreate) {

}
