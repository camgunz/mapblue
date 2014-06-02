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
	Name                string
	Coordinates         string
	RacePopulation      int
	BlackCount          int
	HispanicCount       int
	OtherRaceCount      int
	HouseholdPopulation int
	UnmarriedCount      int
	ChildlessCount      int
}

type CensusBlocks struct {
	Blocks []CensusBlock
}

func (cb *CensusBlock) MarshalJSON() ([]byte, error) {
	return []byte(cb.Coordinates), nil
}

/*
const blockQueryTemplate = "SELECT tabblock_id, name, ST_AsGeoJSON(the_geom) " +
	"FROM tabblock WHERE ST_Intersects(the_geom, ST_GeometryFromText(" +
	"'SRID=4269;MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))'" +
	"));"
*/

const blockQueryTemplate = "SELECT tb.name, ST_AsGeoJSON(tb.the_geom), " +
	"p5.p0050001, p5.p0050004, p5.p0050005, p5.p0050006, " +
	"p5.p0050007, p5.p0050008, p5.p0050009, p5.p0050010, " +
	"p19.p0190001, p19.p0190002, p19.p0190009, p19.p0190010, " +
	"p19.p0190012, p19.p0190013, p19.p0190015, p19.p0190016, " +
	"p19.p0190017 " +
	"FROM tabblock AS tb, geo_locations as gl, p5 as p5, p19 as p19 " +
	"WHERE ST_Intersects(the_geom, ST_GeomFromEWKT(" +
	"'SRID=4269;MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))'" +
	")) " +
	"AND gl.intptlon = tb.intptlon " +
	"AND gl.intptlat = tb.intptlat " +
	"AND p5.logrecno = gl.logrecno " +
	"AND p19.logrecno = gl.logrecno;"

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
		var name, coordinates string
		var racePop, blackCount, aianCount, asianCount, nhopiCount int
		var otherRaceCount, multiracialCount, hispanicCount int
		var householdPop, singleNoFamilyCount, husbandAndWifeChildlessCount int
		var unmarriedWithFamilyCount, singleDadCount int
		var unmarriedMaleFamilyCount, singleMomCount int
		var unmarriedFemaleFamilyCount, nonFamilyCount int

		if blockCount > len(censusBlocks.Blocks) {
			newBlocks := make([]CensusBlock, len(censusBlocks.Blocks)+blockChunkSize)
			copy(newBlocks, censusBlocks.Blocks)
			censusBlocks.Blocks = newBlocks
		}
		block := &censusBlocks.Blocks[blockCount]
		err = blockRows.Scan(
			&name, &coordinates,
			&racePop,
			&blackCount, &aianCount, &asianCount, &nhopiCount,
			&otherRaceCount, &multiracialCount, &hispanicCount,
			&householdPop,
			&singleNoFamilyCount, &husbandAndWifeChildlessCount,
			&unmarriedWithFamilyCount,
			&singleDadCount, &unmarriedMaleFamilyCount,
			&singleMomCount, &unmarriedFemaleFamilyCount,
			&nonFamilyCount,
		)
		if err != nil {
			log.Fatal(err)
		}

		block.Name = name
		block.Coordinates = coordinates
		block.RacePopulation = racePop
		block.BlackCount = blackCount
		block.HispanicCount = hispanicCount
		block.OtherRaceCount = aianCount + asianCount + nhopiCount +
			otherRaceCount + multiracialCount
		block.HouseholdPopulation = householdPop
		block.UnmarriedCount = singleNoFamilyCount +
			unmarriedWithFamilyCount + nonFamilyCount
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
