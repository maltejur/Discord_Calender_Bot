package database

import (
	"database/sql"
	"log"
	"time"

	//Database driver
	_ "github.com/mattn/go-sqlite3"
)

var appointmentdb *sql.DB
var lookupdb *sql.DB

//Appointment is the basic type to store an apointment
type Appointment struct {
	description string
	deadline    time.Time
	ty          string
}

//InitDB opens database from file
func InitDB(appath string, lopath string) {
	dbb, err := sql.Open("sqlite3", appath)
	if err != nil {
		log.Panic(err)
	}
	appointmentdb = dbb

	dbb, err = sql.Open("sqlite3", lopath)
	if err != nil {
		log.Panic(err)
	}
	lookupdb = dbb

	sqlStmt := `CREATE TABLE IF NOT EXISTS "Channels" (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		channelID TEXT NOT NULL,
		channelName TEXT NOT NULL
	);`
	_, err = lookupdb.Exec(sqlStmt)
	if err != nil {
		log.Panic(err)
	}

	sqlStmt = `CREATE TABLE IF NOT EXISTS "Users" (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		userID TEXT NOT NULL,
		userName TEXT NOT NULL
	);`
	_, err = lookupdb.Exec(sqlStmt)
	if err != nil {
		log.Panic(err)
	}
}

//GetAppointmentsFromDatabase recieves all appointments from a given channel
func GetAppointmentsFromDatabase(channelID string) ([]Appointment, error) {
	var aps []Appointment
	sqlStmt := `SELECT * FROM "` + channelID + `";`
	stmt, err := appointmentdb.Prepare(sqlStmt)
	if err != nil {
		return aps, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return aps, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var ap Appointment
		var deadline string
		err = rows.Scan(&id, &ap.description, &deadline, &ap.ty)
		if err != nil {
			return aps, err
		}
		ap.deadline, err = time.Parse(time.UnixDate, deadline)
		if err != nil {
			return aps, err
		}
		aps = append(aps, ap)
	}
	return aps, nil
}

//WriteAppointmentToDatabse writes Appointment to Database
func WriteAppointmentToDatabse(channelID string, ap Appointment) error {
	sqlStmt := `INSERT INTO "` + channelID + `" (
		description,deadline,type
	) VALUES (
		?,?,?
	);`
	stmt, err := appointmentdb.Prepare(sqlStmt)
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
	_, err := appointmentdb.Exec(sqlStmt)
	return err
}
