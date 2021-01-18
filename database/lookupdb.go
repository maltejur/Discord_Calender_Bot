package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

const (
	ChannelC = 0
	UserC    = 1
)

var lookupdb *sql.DB

//ClearLookupDB deletes all tables from Lookupdb
func ClearLookupDB() error {
	sqlStmt := `PRAGMA writable_schema = 1;
	DELETE FROM sqlite_master;
	PRAGMA writable_schema = 0;
	VACUUM;
	PRAGMA integrity_check;`
	_, err := lookupdb.Exec(sqlStmt)
	if err != nil {
		return err
	}
	sqlStmt = `CREATE TABLE IF NOT EXISTS "Channels" (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		channelID TEXT NOT NULL,
		channelName TEXT NOT NULL
	);`
	_, err = lookupdb.Exec(sqlStmt)
	if err != nil {
		return err
	}
	sqlStmt = `CREATE TABLE IF NOT EXISTS "Users" (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		userID TEXT NOT NULL,
		userName TEXT NOT NULL
	);`
	_, err = lookupdb.Exec(sqlStmt)
	return err
}

//NewUserTable creates table for new user in lookupdb
func NewUserTable(userID string) error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS "` + userID + `" (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		channelID TEXT NOT NULL,
		channelName TEXT NOT NULL);`
	_, err := lookupdb.Exec(sqlStmt)
	return err
}

//WriteNewLookupEntry writes new Channels of a  user in table
func WriteNewLookupEntry(userID string, channel Channel) error {

	sqlStmt := fmt.Sprintf(`INSERT INTO "%s" (channelID,channelName) VALUES ("%s","%s");`,
		userID, channel.ID, channel.Name)
	_, err := lookupdb.Exec(sqlStmt)
	return err
}

func readChannels() map[string]struct{} {
	sqlStmt := `SELECT * FROM Channels;`
	channels := make(map[string]struct{})
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
		var channelName string
		var id int
		err = rows.Scan(&id, &channelID, &channelName)
		if err != nil {
			log.Panic(err)
		}
		channels[channelID] = struct{}{}
	}
	return channels
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
		var id int
		var userName string
		err = rows.Scan(&id, &userID, &userName)
		if err != nil {
			log.Panic(err)
		}
		users = append(users, userID)
	}
	return users
}

//PopulateLookups reads the lookup information from the lookup db and stores it into the Lookup variabels in messageprocessing
func PopulateLookups() (map[string]struct{}, map[string][]Channel) {
	channelss := readChannels()
	users := readUsers()
	sqlStmt := `SELECT * FROM "%s"`
	lookup := make(map[string][]Channel)
	for _, user := range users {
		var channels []Channel
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
			var channel Channel
			var id int
			err = rows.Scan(&id, &channel.ID, &channel.Name)
			channels = append(channels, channel)
		}
		lookup[user] = channels
	}
	return channelss, lookup
}

//WriteNewID adds new Channel to Channels Table in the lookupdb
func WriteNewID(ID string, name string, table int) error {
	var sqlStmt string
	if table == ChannelC {
		sqlStmt = fmt.Sprintf(`INSERT INTO Channels (channelID, channelName) VALUES ("%s","%s");`, ID, name)
	} else if table == UserC {
		sqlStmt = fmt.Sprintf(`INSERT INTO Users (userID,userName) VALUES ("%s" ,"%s");`, ID, name)
	} else {
		return errors.New("WriteNewID wrong table")
	}
	_, err := lookupdb.Exec(sqlStmt)
	return err
}
