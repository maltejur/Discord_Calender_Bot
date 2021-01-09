package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/poodlenoodle42/Discord_Calender_Bot/messageprocessing"
)

const (
	Channel = 0
	User    = 1
)

var lookupdb *sql.DB

//WriteNewLookupEntry writes Channels of a new user in table
func WriteNewLookupEntry(userID string, channels []messageprocessing.Channel) error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS "` + userID + `"(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		channelID TEXT NOT NULL,
		channelName TEXT NOT NULL);`
	_, err := lookupdb.Exec(sqlStmt)
	if err != nil {
		return err
	}
	sqlStmt = `INSERT INTO "` + userID + `" (channelID,channelName) VALUES `
	vals := []interface{}{}
	for _, channel := range channels {
		sqlStmt += "(?,?),"
		vals = append(vals, channel.ID, channel.Name)
	}
	stmt, err := lookupdb.Prepare(sqlStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}
	return nil
}

func readChannels() {
	sqlStmt := `SELECT * FROM Channels;`
	var channels []string
	stmt, err := lookupdb.Prepare(sqlStmt)
	if err != nil {
		log.Panic(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var channelID string
		err = rows.Scan(&channelID)
		channels = append(channels, channelID)
	}
	messageprocessing.KnownChannels = channels
}

func readUsers() []string {
	sqlStmt := `SELECT * FROM Users;`
	var users []string
	stmt, err := lookupdb.Prepare(sqlStmt)
	if err != nil {
		log.Panic(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		err = rows.Scan(&userID)
		users = append(users, userID)
	}
	return users
}

//PopulateLookups reads the lookup information from the lookup db and stores it into the Lookup variabels in messageprocessing
func PopulateLookups() {
	readChannels()
	users := readUsers()
	sqlStmt := `SELECT * FROM "%s"`

	for _, user := range users {
		var channels []messageprocessing.Channel
		sqlStmtT := fmt.Sprintf(sqlStmt, user)
		stmt, err := lookupdb.Prepare(sqlStmtT)
		if err != nil {
			log.Panic(err)
		}
		defer stmt.Close()
		rows, err := stmt.Query()
		if err != nil {
			log.Panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var channel messageprocessing.Channel
			err = rows.Scan(&channel.ID, &channel.Name)
			channels = append(channels, channel)
		}
		messageprocessing.Lookup[user] = channels
	}
}

//WriteNewID adds new Channel to Channels Table in the lookupdb
func WriteNewID(ID string, name string, table int) error {
	var sqlStmt string
	if table == Channel {
		sqlStmt = `INSERT INTO Channels (channelID, channelName) VALUES (` + ID + "," + name + `);`
	} else if table == User {
		sqlStmt = `INSERT INTO Users (userID,userName) VALUES (` + ID + "," + name + `);`
	} else {
		return errors.New("WriteNewID wrong table")
	}
	_, err := lookupdb.Exec(sqlStmt)
	return err
}
