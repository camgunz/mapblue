package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

var db *sql.DB

type CensusBlock struct {
	ID          string
	Name        string
	Coordinates string
}

type CensusBlocks struct {
	Blocks []CensusBlock
}

func (cb *CensusBlock) MarshalJSON() ([]byte, error) {
	return []byte(cb.Coordinates), nil
}

const countyQueryTemplate string = "SELECT countyfp FROM tract WHERE " +
	"ST_Intersects(the_geom, ST_GeogFromText('" +
	"MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))" +
	"'))"

const blockQueryTemplate string = "SELECT tabblock_id, name, " +
	"ST_AsGeoJSON(the_geom) " +
	"FROM tabblock WHERE countyfp = '%s' AND ST_Intersects(" +
	"the_geom, ST_GeogFromText('" +
	"MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))" +
	"'))"

const blockChunkSize = 50

func send400(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, msg)
}

func send500(w http.ResponseWriter, e error) {
	log.Print(e)
	w.WriteHeader(http.StatusInternalServerError)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var lat1, lon1, lat2, lon2 string
	form := r.URL.Query()
	if _, ok := form["lat1"]; ok {
		if _, err := strconv.ParseFloat(form["lat1"][0], 64); err != nil {
			send400(w, "Invalid value for lat1")
			return
		}
		lat1 = form["lat1"][0]
	} else {
		send400(w, "Required parameter 'lat1' not given")
		return
	}
	if _, ok := form["lon1"]; ok {
		if _, err := strconv.ParseFloat(form["lon1"][0], 64); err != nil {
			send400(w, "Invalid value for lon1")
			return
		}
		lon1 = form["lon1"][0]
	} else {
		send400(w, "Required parameter 'lon1' not given")
		return
	}
	if _, ok := form["lat2"]; ok {
		if _, err := strconv.ParseFloat(form["lat2"][0], 64); err != nil {
			send400(w, "Invalid value for lat2")
			return
		}
		lat2 = form["lat2"][0]
	} else {
		send400(w, "Required parameter 'lat2' not given")
		return
	}
	if _, ok := form["lon2"]; ok {
		if _, err := strconv.ParseFloat(form["lon2"][0], 64); err != nil {
			send400(w, "Invalid value for lon2")
			return
		}
		lon2 = form["lon2"][0]
	} else {
		send400(w, "Required parameter 'lon2' not given")
		return
	}

	countyQueryString := fmt.Sprintf(
		countyQueryTemplate,
		lat1, lon1, lat2, lon1, lat1, lon2, lat2, lon2, lat1, lon1,
	)
	countyRows, err := db.Query(countyQueryString)
	if err != nil {
		send500(w, err)
		return
	}
	counties := make([]string, 92)
	for countyRows.Next() {
		var county string
		if err = countyRows.Scan(&county); err != nil {
			send500(w, err)
			return
		}
		counties = append(counties, county)
	}
	if err = countyRows.Err(); err != nil {
		send500(w, err)
		return
	}
	if len(counties) == 0 {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		fmt.Fprint(w, "{}")
		return
	}

	censusBlocks := CensusBlocks{}
	censusBlocks.Blocks = make([]CensusBlock, 50)
	blockCount := 0
	for _, county := range counties {
		blockQueryString := fmt.Sprintf(
			blockQueryTemplate,
			county, lat1, lon1, lat2, lon1, lat1, lon2, lat2, lon2, lat1, lon1,
		)
		blockRows, err := db.Query(blockQueryString)
		if err != nil {
			send500(w, err)
			return
		}
		for blockRows.Next() {
			if blockCount > len(censusBlocks.Blocks) {
				newBlocks := make([]CensusBlock, len(censusBlocks.Blocks)+50)
				copy(newBlocks, censusBlocks.Blocks)
				censusBlocks.Blocks = newBlocks
			}
			block := &censusBlocks.Blocks[blockCount]
			err = blockRows.Scan(&block.ID, &block.Name, &block.Coordinates)
			if err != nil {
				log.Fatal(err)
			}
			blockCount++
		}
		if err = blockRows.Err(); err != nil {
			send500(w, err)
			return
		}
	}
	censusBlocks.Blocks = censusBlocks.Blocks[:blockCount]

	jsonData, err := json.Marshal(censusBlocks)
	if err != nil {
		send500(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	fmt.Fprint(w, string(jsonData))
}

func main() {
	pg, err := sql.Open(
		"postgres",
		"host=/var/run/postgresql dbname=census "+
			"user=census sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}
	db = pg

	http.HandleFunc("/", handler)
	fmt.Println("Listening on 0.0.0.0:8080")
	http.ListenAndServe(":8080", nil)
	err = db.Close()
	fmt.Println("Closing DB")
	if err != nil {
		log.Fatal(err)
	}
}
