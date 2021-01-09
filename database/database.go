package database

import (
	"database/sql"
	"log"
	"time"

	//Database driver
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

//Appointment is the basic type to store an apointment
type Appointment struct {
	description string
	deadline    time.Time
	ty          string
}

//InitDB opens database from file
func InitDB(path string) {
	dbb, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panic(err)
	}
	db = dbb
}

//GetAppointmentsFromDatabase recieves all appointments from a given channel
func GetAppointmentsFromDatabase(channelID string) []Appointment {

}

//WriteAppointmentToDatabse writes Appointment to Database
func WriteAppointmentToDatabse(channelID string, ap Appointment) error {

}
