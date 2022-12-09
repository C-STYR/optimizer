package optimizerdb

import (
	"database/sql"
	"log"

	"github.com/fatih/color"
	"github.com/mattn/go-sqlite3"
)


/*
	Creates a table 'hits' in the database of every filename
	with columns id, name and count
*/
func TryCreate(db *sql.DB) bool {
	_, err := db.Exec(`
		CREATE TABLE hits (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE,
			count INTEGER
		);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			} else {
				return false
			}
		} else {
			log.Fatal(err)
		}
	}
	return true
}

/*
	Increments the count value for a given name
	If name not found, add name and set hit to 1
*/

func IncrementHitCount(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO
		hits(name, count) VALUES(?, 1)
	ON CONFLICT(name) DO UPDATE SET count = count + 1`, name)

	if err != nil {
		display := color.New(color.FgRed).SprintFunc()
		log.Printf("%s\n", display(err))
		return err
	}
	return nil
}