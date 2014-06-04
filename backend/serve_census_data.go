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
	Name      string
	Geometry  map[string]interface{}
	Over18    int
	Black     int
	Hispanic  int
	OtherRace int
	Unmarried int
	Childless int
}

type CensusBlocks struct {
	Blocks []CensusBlock
}

const blockQueryTemplate = "SELECT tb.name, ST_AsGeoJSON(tb.the_geom), " +
	"p11.p0110006, p11.p0110007, p11.p0110008, p11.p0110009, p11.p0110010, " +
	"p11.p0110011, p11.p0110002, " +
	"p16.p0160003, " +
	"p19.p0190009, p19.p0190013, p19.p0190016, " +
	"p29.p0290007, p29.p0290015, p29.p0290018 " +
	"FROM tabblock AS tb, geo_locations as gl, p11, p16, p19, p29 " +
	"WHERE ST_Intersects(the_geom, ST_GeomFromEWKT(" +
	"'SRID=4269;MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))'" +
	")) " +
	"AND gl.intptlon = tb.intptlon " +
	"AND gl.intptlat = tb.intptlat " +
	"AND p11.logrecno = gl.logrecno " +
	"AND p16.logrecno = gl.logrecno " +
	"AND p19.logrecno = gl.logrecno " +
	"AND p29.logrecno = gl.logrecno;"

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
		var name, geoJSONData string
		var blacks, aians, asians, nhopis, others, multis, hispanics int
		var over18 int
		var childlessHusbandAndWifeFamilies int
		var childlessMaleFamilies int
		var childlessFemaleFamilies int
		var spouses int
		var sonsOrDaughtersInLaw int
		var unrelatedRoommates int

		if blockCount > len(censusBlocks.Blocks) {
			newBlocks := make([]CensusBlock, len(censusBlocks.Blocks)+blockChunkSize)
			copy(newBlocks, censusBlocks.Blocks)
			censusBlocks.Blocks = newBlocks
		}
		err = blockRows.Scan(
			&name,
			&geoJSONData,
			&blacks,
			&aians,
			&asians,
			&nhopis,
			&others,
			&multis,
			&hispanics,
			&over18,
			&childlessHusbandAndWifeFamilies,
			&childlessMaleFamilies,
			&childlessFemaleFamilies,
			&spouses,
			&sonsOrDaughtersInLaw,
			&unrelatedRoommates,
		)
		if err != nil {
			send500(w, err)
			return
		}

		var geoJSON interface{}
		err = json.Unmarshal([]byte(geoJSONData), &geoJSON)
		if err != nil {
			send500(w, err)
			return
		}

		block := &censusBlocks.Blocks[blockCount]
		block.Name = name
		block.Geometry = geoJSON.(map[string]interface{})
		block.Over18 = over18
		block.Black = blacks
		block.Hispanic = hispanics
		block.OtherRace = aians + asians + nhopis + others + multis
		block.Unmarried = over18 - ((spouses * 2) + (sonsOrDaughtersInLaw * 2))
		block.Childless = unrelatedRoommates +
			(childlessHusbandAndWifeFamilies * 2) +
			childlessMaleFamilies +
			childlessFemaleFamilies

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
	jsonData, err := json.MarshalIndent(censusBlocks, "", "    ")
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
