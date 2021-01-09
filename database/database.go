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
	var aps []Appointment
	sqlStmt := `SELECT * FROM "` + channelID + `";`
	stmt, err := db.Prepare(sqlStmt)

}

//WriteAppointmentToDatabse writes Appointment to Database
func WriteAppointmentToDatabse(channelID string, ap Appointment) error {
	sqlStmt := `INSERT INTO "` + channelID + `" (
		description,deadline,type
	) VALUES (
		?,?,?
	);`
	stmt, err := db.Prepare(sqlStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(ap.description, ap.deadline.String(), ap.ty)
	return err
}

//MakeNewChannelTable creates new table for new Channel
func MakeNewChannelTable(channelID string) error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS "` + channelID + `"(
		id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		description TEXT,
		deadline TEXT NOT NULL,
		type TEXT NOT NULL);`
	_, err := db.Exec(sqlStmt)
	return err
}
