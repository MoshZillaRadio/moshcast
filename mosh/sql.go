package mosh

import (
	"database/sql"
	"moshcast/utils"

	"strings"
	"time"

	_ "github.com/MoshZillaRadio/go-sqlite3-ext"
	"golang.org/x/crypto/bcrypt"
)

type SQL struct {
	DBFile string
	logger Logger
	DB     *sql.DB
}

type MetaEntry struct {
	Title string
	Date  string
}

func (s SQL) Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite3-ext", s.DBFile)

	if err != nil {
		s.logger.Error("%s", err)
	}
	return db, err
}

func (s *SQL) CreateTable(db *sql.DB) {
	stmt := `CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT UNIQUE,
                password TEXT,
				salt TEXT
        );
        CREATE TABLE IF NOT EXISTS meta (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                title TEXT,
                date DATETIME DEFAULT CURRENT_TIMESTAMP
        );`
	_, err := db.Exec(stmt)
	if err != nil {
		s.logger.Error("%s", err)
	}
}

func (s *SQL) InsertUser(db *sql.DB, username, password string) {
	salt := utils.GenerateSalt()
	password = utils.GeneratePasswordHash(username, password, salt)

	_, err := db.Exec("INSERT INTO users (username, password, salt) VALUES (?, ?, ?)", username, password)
	if err != nil {
		s.logger.Error("%s", err)
	}
}

func (s *SQL) InsertMetaData(db *sql.DB, data string) {
	currentTime := time.Now().Format("2006-01-02 15:04")

	_, err := db.Exec("INSERT INTO meta (title, date) VALUES (?, ?)", data, currentTime)
	if err != nil {
		s.logger.Error("%s", err)
	}
}

func (s *SQL) SelectMetaData(db *sql.DB, date string) []MetaEntry {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT title, date FROM meta WHERE date LIKE '")
	queryBuilder.WriteString(date + "%")
	queryBuilder.WriteString("' ORDER BY date DESC LIMIT 5")
	query := queryBuilder.String()

	if utils.CheckUserInput(date) {
		return []MetaEntry{{Title: "", Date: ""}}
	}

	rows, err := db.Query(query)
	if err != nil {
		s.logger.Error(err.Error())
		return []MetaEntry{{Title: "", Date: ""}}
	}
	defer rows.Close()

	var results []MetaEntry
	for rows.Next() {
		var meta MetaEntry
		if err := rows.Scan(&meta.Title, &meta.Date); err != nil {
			s.logger.Error("Row scan error: " + err.Error())
			continue
		}
		results = append(results, meta)
	}
	return results
}

func (s *SQL) Authentication(db *sql.DB, username, password string) bool {
	var storedHash, storedSalt string

	stmt, err := db.Prepare("SELECT password, salt FROM users WHERE username = ?")
	if err != nil {
		s.logger.Error("Failed to prepare statement: %v", err)
		return false
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&storedHash, &storedSalt)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Log("User not found")
		} else {
			s.logger.Error("Database query error: %v", err)
		}
		return false
	}

	password = utils.GeneratePasswordString(username, password, storedSalt)
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
}
