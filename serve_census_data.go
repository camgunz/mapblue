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

const blockQueryTemplate = "SELECT tabblock_id, name, ST_AsGeoJSON(the_geom) " +
	"FROM tabblock WHERE ST_Intersects(the_geom, ST_GeometryFromText(" +
	"'SRID=4269;MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))'" +
	"));"

const blockChunkSize = 3000

func send400(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, msg)
}

func send500(w http.ResponseWriter, e error) {
	log.Print(e)
	w.WriteHeader(http.StatusInternalServerError)
}

func checkParam(w http.ResponseWriter, r *http.Request, paramName string) {
	form := r.URL.Query()
	if _, ok := form[paramName]; ok {
		if _, err := strconv.ParseFloat(form[paramName][0], 64); err != nil {
			send400(w, fmt.Sprintf("Invalid value for %s", paramName))
			return
		}
	} else {
		send400(w, fmt.Sprintf("Required parameter '%s' not given", paramName))
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	form := r.URL.Query()

	checkParam(w, r, "lat1")
	checkParam(w, r, "lon1")
	checkParam(w, r, "lat2")
	checkParam(w, r, "lon2")

	lat1 := form["lat1"][0]
	lon1 := form["lon1"][0]
	lat2 := form["lat2"][0]
	lon2 := form["lon2"][0]

	censusBlocks := CensusBlocks{}
	censusBlocks.Blocks = make([]CensusBlock, blockChunkSize)
	blockCount := 0
	blockRows, err := db.Query(fmt.Sprintf(blockQueryTemplate,
		lon1, lat1, lon2, lat1, lon2, lat2, lon2, lat2, lon1, lat1,
	))
	if err != nil {
		send500(w, err)
		return
	}
	for blockRows.Next() {
		if blockCount > len(censusBlocks.Blocks) {
			newBlocks := make([]CensusBlock, len(censusBlocks.Blocks)+blockChunkSize)
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

	if blockCount == 0 {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		fmt.Fprint(w, "{}")
		return
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
		"host=/var/run/postgresql dbname=census user=census sslmode=disable",
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
