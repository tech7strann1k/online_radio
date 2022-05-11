package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type Stream struct {
	Id                     int
	Title, Land, Logo, Url string
}

type StreamDatabase struct {
	*sql.DB
}

func InitDB(file string) *StreamDatabase {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		panic(err)
	}
	streamDatabase := &StreamDatabase{DB: db}
	return streamDatabase

}

func (db *StreamDatabase) itemExists(arr []string, item interface{}) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func (db *StreamDatabase) LoadLandList() []string {
	defer db.Close()
	rows, err := db.Query("SELECT land FROM radiometadata")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	landList := []string{}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			fmt.Println(err)
		}
		if !db.itemExists(landList, s) {
			landList = append(landList, s)
		}
	}
	return landList
}

func (db *StreamDatabase) LoadData(argv interface{}) []Stream {
	defer db.Close()
	var query = `SELECT * FROM radiometadata WHERE land LIKE "%Russia%"`
	// if argv != nil {
	// 	query += "WHERE land LIKE ?"
	// }
	rows, err := db.Query(query, argv)
	if err != nil {
		panic(err)
	}
	streamList := []Stream{}
	defer rows.Close()
	for rows.Next() {
		var s Stream
		err := rows.Scan(&s.Id, &s.Title, &s.Land, &s.Logo, &s.Url)
		if err != nil {
			fmt.Println(err)
		}
		streamList = append(streamList, s)
	}
	return streamList

}

func (db *StreamDatabase) AddData(nameOfStream, urlOfStream, iconOfSstream string) {
	defer db.Close()
	var query = "INSERT INTO favourites VALUES"
	_, err := db.Query(query, nameOfStream, urlOfStream, iconOfSstream)
	if err != nil {
		panic(err)
	}
}