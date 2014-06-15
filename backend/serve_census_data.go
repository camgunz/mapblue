package main

import (
	"compress/gzip"
	"compress/zlib"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CensusBlockProperties struct {
	Name      string `json:"name"`
	Over18    int    `json:"over18"`
	Black     int    `json:"black"`
	Hispanic  int    `json:"hispanic"`
	OtherRace int    `json:"otherRace"`
	Unmarried int    `json:"unmarried"`
	Childless int    `json:"childless"`
}

type CensusBlock struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties CensusBlockProperties  `json:"properties"`
}

type CensusBlocks struct {
	Type     string        `json:"type"`
	Features []CensusBlock `json:"features"`
}

const postgresAddress = "/var/run/postgresql"
const postgresUser = "census"
const postgresDatabase = "census"
const postgresSSLMode = "disable"

// const hostAddressAndPort = "127.0.0.1:8080"
const hostAddressAndPort = "0.0.0.0:8080"
const blockChunkSize = 3000
const blockQueryTemplate = "SELECT tb.tabblock_id, tb.name, " +
	"ST_AsGeoJSON(tb.the_geom), " +
	"p11.p0110006, p11.p0110007, p11.p0110008, p11.p0110009, p11.p0110010, " +
	"p11.p0110011, p11.p0110002, p16.p0160003, p19.p0190009, p19.p0190013, " +
	"p19.p0190016, p29.p0290007, p29.p0290015, p29.p0290018 " +
	"FROM tabblock AS tb, geo_locations as gl, p11, p16, p19, p29 " +
	"WHERE ST_Intersects(the_geom, ST_GeomFromEWKT(" +
	"'SRID=4269;MULTIPOLYGON(((%s %s, %s %s, %s %s, %s %s, %s %s)))'" +
	")) " +
	"AND gl.sumlev IN ('101', '750', '755') " +
	"AND gl.intptlon = tb.intptlon AND gl.intptlat = tb.intptlat " +
	"AND p11.logrecno = gl.logrecno AND p16.logrecno = gl.logrecno " +
	"AND p19.logrecno = gl.logrecno AND p29.logrecno = gl.logrecno;"

var db *sql.DB

func send400(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, msg)
}

func send500(w http.ResponseWriter, e error) {
	log.Print(e)
	w.WriteHeader(http.StatusInternalServerError)
}

func checkParam(w http.ResponseWriter, r *http.Request, paramName string) bool {
	form := r.URL.Query()
	if _, ok := form[paramName]; ok {
		if _, err := strconv.ParseFloat(form[paramName][0], 64); err != nil {
			send400(w, fmt.Sprintf("Invalid value for %s", paramName))
			return false
		}
	} else {
		send400(w, fmt.Sprintf("Required parameter '%s' not given", paramName))
		return false
	}

	return true
}

func lookup(w http.ResponseWriter, r *http.Request) {
	var supportedEncodings = r.Header.Get("Accept-Encoding")
	var supportsGZIP = strings.Contains(supportedEncodings, "gzip")
	var supportsZLIB = strings.Contains(supportedEncodings, "deflate")
	var lat1, lon1, lat2, lon2 string
	form := r.URL.Query()

	if checkParam(w, r, "lat1") && checkParam(w, r, "lon1") &&
		checkParam(w, r, "lat2") && checkParam(w, r, "lon2") {
		lat1 = form["lat1"][0]
		lon1 = form["lon1"][0]
		lat2 = form["lat2"][0]
		lon2 = form["lon2"][0]
	} else {
		return
	}

	censusBlocks := CensusBlocks{}
	censusBlocks.Type = "FeatureCollection"
	censusBlocks.Features = make([]CensusBlock, blockChunkSize)
	blockCount := 0

	blockRows, err := db.Query(fmt.Sprintf(blockQueryTemplate,
		lon1, lat1, lon2, lat1, lon2, lat2, lon1, lat2, lon1, lat1,
	))
	if err != nil {
		send500(w, err)
		return
	}

	for blockRows.Next() {
		var (
			blockID, name, geoJSONData string
			blacks, aians, asians, nhopis, others, multis, hispanics, over18,
			childlessHusbandAndWifeFamilies, childlessMaleFamilies,
			childlessFemaleFamilies, spouses, sonsOrDaughtersInLaw,
			unrelatedRoommates int
			geoJSON interface{}
		)

		if blockCount >= len(censusBlocks.Features) {
			newBlocks := make(
				[]CensusBlock, len(censusBlocks.Features)+blockChunkSize,
			)
			copy(newBlocks, censusBlocks.Features)
			censusBlocks.Features = newBlocks
		}
		err = blockRows.Scan(
			&blockID, &name, &geoJSONData, &blacks, &aians, &asians, &nhopis,
			&others, &multis, &hispanics, &over18,
			&childlessHusbandAndWifeFamilies, &childlessMaleFamilies,
			&childlessFemaleFamilies, &spouses, &sonsOrDaughtersInLaw,
			&unrelatedRoommates,
		)
		if err != nil {
			send500(w, err)
			return
		}

		err = json.Unmarshal([]byte(geoJSONData), &geoJSON)
		if err != nil {
			send500(w, err)
			return
		}

		block := &censusBlocks.Features[blockCount]
		block.ID = blockID
		block.Type = "Feature"
		block.Geometry = geoJSON.(map[string]interface{})
		block.Properties.Name = name
		block.Properties.Over18 = over18
		block.Properties.Black = blacks
		block.Properties.Hispanic = hispanics
		block.Properties.OtherRace =
			aians + asians + nhopis + others + multis
		block.Properties.Unmarried =
			over18 - ((spouses * 2) + (sonsOrDaughtersInLaw * 2))
		block.Properties.Childless = unrelatedRoommates +
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

	censusBlocks.Features = censusBlocks.Features[:blockCount]
	jsonData, err := json.MarshalIndent(censusBlocks, "", "    ")
	if err != nil {
		send500(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	if supportsGZIP {
		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		gzipWriter.Write(jsonData)
		gzipWriter.Close()
		return
	}
	if supportsZLIB {
		w.Header().Set("Content-Encoding", "deflate")
		zlibWriter, _ := zlib.NewWriterLevel(w, zlib.BestSpeed)
		zlibWriter.Write(jsonData)
		zlibWriter.Close()
		return
	}

	fmt.Fprint(w, string(jsonData))
}

func main() {
	pg, err := sql.Open(
		"postgres", fmt.Sprintf("host=%s dbname=%s user=%s sslmode=%s",
			postgresAddress, postgresDatabase, postgresUser, postgresSSLMode,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	db = pg

	http.HandleFunc("/", lookup)
	fmt.Printf("Listening on %s\n", hostAddressAndPort)
	http.ListenAndServe(hostAddressAndPort, nil)
	if err = db.Close(); err != nil {
		log.Fatal(err)
	}
}
