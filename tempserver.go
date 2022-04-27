package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/gorilla/mux"
	"github.com/magiconair/properties"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

var db *sql.DB
var cfg Config

type Config struct {
	DbFile string `properties:"dbfile"`
}

func generateLineItems(name int, count int) []opts.LineData { //nolint:unparam
	items := make([]opts.LineData, 0)

	//	items := make([]Reading, 0)

	rows, err := db.Query("SELECT Id FROM reading ORDER BY ID DESC LIMIT 1")
	Check(err)
	defer rows.Close()
	lastid := 0
	for rows.Next() {
		read := Reading{}
		if err = rows.Scan(&read.Id); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		lastid = read.Id
	}
	rows, err = db.Query("SELECT Sensor, Temperature FROM reading  WHERE ID >= $1  AND Sensor = $2 LIMIT $3", lastid-count, name, count)

	//	rows, err := db.Query("SELECT Sensor, Temperature FROM reading ORDER BY ID DESC LIMIT $1 WHERE Sensor = $2", count, name)
	Check(err)
	defer rows.Close()
	for rows.Next() {
		read := Reading{}
		// create an instance of `Bird` and write the result of the current row into it
		if err = rows.Scan(&read.Sensor, &read.Temperature); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		//fmt.Printf("found read: %+v\n", read)
		// append the current instance to the slice of birds
		items = append(items, opts.LineData{Value: read.Temperature})
	}
	return items
}

func httpserver(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Line example in Westeros theme",
			Subtitle: "Line chart rendered by the http server this time",
		}))

	// Put data into instance
	line.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("28-1b61221e64ff", generateLineItems(1, 100)).
		AddSeries("28-7167221e64ff", generateLineItems(2, 100)).
		AddSeries("28-7167221e64ff", generateLineItems(3, 100)).
		AddSeries("28-7167221e64ff", generateLineItems(4, 100)).
		AddSeries("28-7167221e64ff", generateLineItems(5, 100)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	err := line.Render(w)
	Check(err)
}

func initdb() {
	var err error
	db, err = sql.Open("sqlite3", cfg.DbFile)
	Check(err)
}

type Reading struct {
	Id          int       `json:"id"`
	Sensor      int       `json:"sensor"`
	TimeStamp   time.Time `json:"timeStamp"`
	Temperature float64   `json:"temperature"`
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func returnLast(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnLast")

	items := make([]Reading, 0)

	rows, err := db.Query("SELECT Id, Sensor, Temperature, Datetime FROM reading ORDER BY ID DESC LIMIT $1", 50)
	Check(err)
	defer rows.Close()
	for rows.Next() {
		read := Reading{}
		timestring := ""
		// create an instance of `Bird` and write the result of the current row into it
		if err = rows.Scan(&read.Id, &read.Sensor, &read.Temperature, &timestring); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		layout := "2006-01-02T15:04:05Z"
		temp, err2 := time.Parse(layout, timestring)
		Check(err2)
		read.TimeStamp = temp

		fmt.Printf("found read: %+v + %s\n", read, timestring)
		// append the current instance to the slice of birds
		items = append(items, read)
	}

	err = json.NewEncoder(w).Encode(items)
	Check(err)
}

func getStart(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: getStart")

	eventID := mux.Vars(r)["id"]

	items := make([]Reading, 0)

	rows, err := db.Query("SELECT Id, Sensor, Temperature, Datetime FROM reading  WHERE ID >= ?  LIMIT $1", eventID, 50)
	Check(err)
	defer rows.Close()
	for rows.Next() {
		read := Reading{}
		timestring := ""
		// create an instance of `Bird` and write the result of the current row into it
		if err = rows.Scan(&read.Id, &read.Sensor, &read.Temperature, &timestring); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		layout := "2006-01-02T15:04:05Z"
		temp, err2 := time.Parse(layout, timestring)
		Check(err2)
		read.TimeStamp = temp.Add(time.Hour * 2)

		fmt.Printf("found read: %+v + %s\n", read, timestring)
		// append the current instance to the slice of birds
		items = append(items, read)
	}

	err = json.NewEncoder(w).Encode(items)
	Check(err)
}

func main() {
	propPtr := flag.String("prop", "tempserver.properties", "a string")
	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	initdb()

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")

	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", httpserver)
	myRouter.HandleFunc("/last", returnLast)
	myRouter.HandleFunc("/start/{id}", getStart).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", myRouter))
}
