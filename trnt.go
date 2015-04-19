package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type TRNT struct {
	id, size_sent, size_total int
	filename, from, to        string
	created_at, updated_at    []byte
}

func Torrents(query string) func() map[int]*TRNT {
	/*
	   Open Database
	*/
	db, err := sql.Open("mysql", "root@/attic")
	if err != nil {
		panic(err.Error())
	}

	/*
	   Query Database

	   | id(int) | from(string) | filename(string) | to(string) | size_sent(int) |
	       | size_total(int) | created_at([]byte) | updated_at([]byte) |

	   INFO: 'size_sent' and 'size_total' are set to -1 if they have not been transfered yet
	   TODO: Change 'created_at' and 'updated_at' type from []byte to time.TIME

	*/
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	return func() map[int]*TRNT {
		/*
		   Closure

		   INFO: Makes a map of torrents from the provided query
		   RETURN: map[int]*TNRT
		*/
		defer db.Close()
		defer rows.Close()
		m := make(map[int]*TRNT)
		var filename, from, to string
		var id, size_sent, size_total int
		var created_at, updated_at []byte
		for rows.Next() {
			if err := rows.Scan(&id, &from, &filename, &to, &size_sent, &size_total, &created_at, &updated_at); err != nil {
				log.Fatal(err)
			}
			tor := new(TRNT)
			tor.WriteTorrent(id, size_sent, size_total, filename, from, to, created_at, updated_at)
			m[id] = tor
		}
		return m
	}
}

func (tor *TRNT) WriteTorrent(id, size_sent, size_total int, filename, from, to string, created_at, updated_at []byte) error {
	/*
	   Write Torrent

	   INFO: Overwrites values for a given torrent
	   RETURN: error
	*/
	tor.id = id
	tor.filename = filename
	tor.from = from
	tor.to = to
	tor.size_sent = size_sent
	tor.size_total = size_total
	tor.created_at = created_at
	tor.updated_at = updated_at
	return nil
}

func main() {

	//allTorrents := Torrents("SELECT * FROM transfers")
	//all := allTorrents()
	unsentTorrents := Torrents("SELECT * FROM transfers WHERE size_sent = -1")
	unsent := unsentTorrents()

	//print the filename of each row
	for k, v := range unsent {
		fmt.Println("Key:", k, " Value:", *v)
	}

}
