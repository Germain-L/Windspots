package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

// Spot Models the data from db
type Spot struct {
	UID  string
	Name string
	Lat  float64
	Lon  float64
}

// Spots is a list of spot
type Spots []Spot

var mainDb *sql.DB

func main() {
	db, errOpenDb := sql.Open("sqlite3", "windspots.db")
	checkErr(errOpenDb)
	mainDb = db

	r := pat.New()
	r.Get("/spots", http.HandlerFunc(getAllSpots))
	r.Post("/spot", http.HandlerFunc(insertSpot))
	r.Get("/spot/:name", http.HandlerFunc(getSpotByName))

	http.Handle("/", r)
	log.Print("Running on 8080")
	err := http.ListenAndServe(":8080", nil)
	checkErr(err)
}

func getAllSpots(w http.ResponseWriter, r *http.Request) {
	rows, err := mainDb.Query("SELECT * FROM spots")
	checkErr(err)

	var spots Spots
	for rows.Next() {
		var spot Spot
		err = rows.Scan(&spot.UID, &spot.Name, &spot.Lat, &spot.Lon)
		checkErr(err)
		spots = append(spots, spot)
	}

	jsonB, errMarshal := json.Marshal(spots)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func insertSpot(w http.ResponseWriter, r *http.Request) {
	var newSpot Spot
	reqBody, err := ioutil.ReadAll(r.Body)
	checkErr(err)

	json.Unmarshal(reqBody, &newSpot)
	stmt, err := mainDb.Prepare("INSERT INTO spots(name, lat, lon) VALUES(?, ?, ?)")
	checkErr(err)

	_, err = stmt.Exec(newSpot.Name, newSpot.Lat, newSpot.Lon)
	checkErr(err)

	w.WriteHeader(http.StatusCreated)
}

func getSpotByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")

	stmt, err := mainDb.Prepare("SELECT * FROM spots where name LIKE ?")
	checkErr(err)

	rows, errQuery := stmt.Query(name)
	checkErr(errQuery)

	var selectedSpot Spot

	for rows.Next() {
		err = rows.Scan(&selectedSpot.UID, &selectedSpot.Name, &selectedSpot.Lat, &selectedSpot.Lon)
		checkErr(err)
	}

	jsonB, errMarshal := json.Marshal(selectedSpot)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
