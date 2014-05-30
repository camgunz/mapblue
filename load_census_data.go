package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type XMLAPIConcepts struct {
	XMLName  xml.Name         `xml:apivariables"`
	Concepts []*XMLAPIConcept `xml:"concept"`
}

type XMLAPIConcept struct {
	XMLname   xml.Name          `xml:"concept"`
	Name      string            `xml:"name,attr"`
	Variables []*XMLAPIVariable `xml:"variable"`
}

type XMLAPIVariable struct {
	XMLName     xml.Name `xml:"variable"`
	Name        string   `xml:"name,attr"`
	Description string   `xml:",chardata"`
}

type APIConcept struct {
	Name          string
	Description   string
	VariableCount int
	Variables     []APIVariable
}

type APIVariable struct {
	Name        string
	Description string
}

type CensusDataFile struct {
	Lock *sync.Mutex
	File *os.File
}

type CensusDataLocation struct {
	DataFile     CensusDataFile
	ColumnOffset int
	ColumnCount  int
}

type CensusTable struct {
	DataLocation CensusDataLocation
	Name         string
	Description  string
	RowCount     int
	Columns      []CensusColumn
}

type CensusColumn struct {
	Name  string
	Value string
}

type GeoLocation struct {
	Fields []GeoLocationField
}

type GeoLocationField struct {
	Name          string
	ReferenceName string
	Size          int
	Position      int
	Numeric       bool
	Value         string
}

type GeoLocationFieldDescription struct {
	Name          string
	ReferenceName string
	Size          int
	Position      int
	Numeric       bool
}

var DB *sql.DB
var CENSUS_DATA_FILES = map[string]CensusDataFile{}
var MAX_DB_CONNECTIONS int = 95
var PRINT_SQL_QUERIES bool = false

var DATA_FILE_TEMPLATE = "in000%02d2010.sf1"
var GEO_FILE = "ingeo2010.sf1"
var PACKING_LIST_FILE = "in2010.sf1.prd.packinglist.txt"
var DATA_DESCRIPTION_URL = "http://www.census.gov/developers/data/sf1.xml"
var CONCEPT_REGEXP = regexp.MustCompile(`(.*)\.(.*)\[(\d\d*)\]`)
var DATA_LOOKUP_REGEXP = regexp.MustCompile(`(.*)\|(\d\d*):(\d\d*)\|$`)
var REQUIRED_FILES = []string{
	"in000012010.sf1",
	"in000022010.sf1",
	"in000032010.sf1",
	"in000042010.sf1",
	"in000052010.sf1",
	"in000062010.sf1",
	"in000072010.sf1",
	"in000082010.sf1",
	"in000092010.sf1",
	"in000102010.sf1",
	"in000112010.sf1",
	"in000122010.sf1",
	"in000132010.sf1",
	"in000142010.sf1",
	"in000152010.sf1",
	"in000162010.sf1",
	"in000172010.sf1",
	"in000182010.sf1",
	"in000192010.sf1",
	"in000202010.sf1",
	"in000212010.sf1",
	"in000222010.sf1",
	"in000232010.sf1",
	"in000242010.sf1",
	"in000252010.sf1",
	"in000262010.sf1",
	"in000272010.sf1",
	"in000282010.sf1",
	"in000292010.sf1",
	"in000302010.sf1",
	"in000312010.sf1",
	"in000322010.sf1",
	"in000332010.sf1",
	"in000342010.sf1",
	"in000352010.sf1",
	"in000362010.sf1",
	"in000372010.sf1",
	"in000382010.sf1",
	"in000392010.sf1",
	"in000402010.sf1",
	"in000412010.sf1",
	"in000422010.sf1",
	"in000432010.sf1",
	"in000442010.sf1",
	"in000452010.sf1",
	"in000462010.sf1",
	"in000472010.sf1",
	"in2010.sf1.prd.packinglist.txt",
	"ingeo2010.sf1",
}

var GeoLocationFieldDescriptions = []GeoLocationFieldDescription{
	{"File Identification", "FILEID", 6, 1, false},
	{"State/U.S. Abbreviation", "STUSAB", 2, 7, false},
	{"Summary Level", "SUMLEV", 3, 9, false},
	{"Geographic Component", "GEOCOMP", 2, 12, false},
	{"Characteristic Iteration", "CHARITER", 3, 14, false},
	{"Characteristic Iteration File Sequence Number", "CIFSN", 2, 17, false},
	{"Logical Record Number", "LOGRECNO", 7, 19, false},
	{"Region", "REGION", 1, 26, false},
	{"Division", "DIVISION", 1, 27, false},
	{"State", "STATE", 2, 28, false},
	{"County", "COUNTY", 3, 30, false},
	{"County Class Code", "COUNTYCC", 2, 33, false},
	{"County Size Code", "COUNTYSC", 2, 35, false},
	{"County Subdivision", "COUSUB", 5, 37, false},
	{"County Subdivision Class Code", "COUSUBCC", 2, 42, false},
	{"County Subdivision Size Code", "COUSUBSC", 2, 44, false},
	{"Place", "PLACE", 5, 46, false},
	{"Place Class Code", "PLACECC", 2, 51, false},
	{"Place Size Code", "PLACESC", 2, 53, false},
	{"Census Tract", "TRACT", 6, 55, false},
	{"Block Group", "BLKGRP", 1, 61, false},
	{"Block", "BLOCK", 4, 62, false},
	{"Internal Use Code", "IUC", 2, 66, false},
	{"Consolidated City", "CONCIT", 5, 68, false},
	{"Consolidated City Class Code", "CONCITCC", 2, 73, false},
	{"Consolidated City Size Code", "CONCITSC", 2, 75, false},
	{"American Indian Area/Alaska Native Area/Hawaiian Home Land (Census)", "AIANHH", 4, 77, false},
	{"American Indian Area/Alaska Native Area/Hawaiian Home Land", "AIANHHFP", 5, 81, false},
	{"American Indian Area/Alaska Native Area/Hawaiian Home Land Class Code", "AIANHHCC", 2, 86, false},
	{"American Indian Trust Land/Hawaiian Home Land Indicator", "AIHHTLI", 1, 88, false},
	{"American Indian Tribal Subdivision (Census)", "AITSCE", 3, 89, false},
	{"American Indian Tribal Subdivision", "AITS", 5, 92, false},
	{"American Indian Tribal Subdivision Class Code", "AITSCC", 2, 97, false},
	{"Tribal Census Tract", "TTRACT", 6, 99, false},
	{"Tribal Block Group", "TBLKGRP", 1, 105, false},
	{"Alaska Native Regional Corporation", "ANRC", 5, 106, false},
	{"Alaska Native Regional Corporation Class Code", "ANRCCC", 2, 111, false},
	{"Metropolitan Statistical Area/Micropolitan Statistical Area", "CBSA", 5, 113, false},
	{"Metropolitan Statistical Area/Micropolitan Statistical Area Size Code", "CBSASC", 2, 118, false},
	{"Metropolitan Division", "METDIV", 5, 120, false},
	{"Combined Statistical Area", "CSA", 3, 125, false},
	{"New England City and Town Area", "NECTA", 5, 128, false},
	{"New England City and Town Area Size Code", "NECTASC", 2, 133, false},
	{"New England City and Town Area Division", "NECTADIV", 5, 135, false},
	{"Combined New England City and Town Area", "CNECTA", 3, 140, false},
	{"Metropolitan Statistical Area/Micropolitan Statistical Area Principal City Indicator", "CBSAPCI", 1, 143, false},
	{"New England City and Town Area Principal City Indicator", "NECTAPCI", 1, 144, false},
	{"Urban Area", "UA", 5, 145, false},
	{"Urban Area Size Code", "UASC", 2, 150, false},
	{"Urban Area Type", "UATYPE", 1, 152, false},
	{"Urban/Rural", "UR", 1, 153, false},
	{"Congressional District (111th)", "CD", 2, 154, false},
	{"State Legislative District (Upper Chamber) (Year 1)", "SLDU", 3, 156, false},
	{"State Legislative District (Lower Chamber) (Year 1)", "SLDL", 3, 159, false},
	{"Voting District", "VTD", 6, 162, false},
	{"Voting District Indicator", "VTDI", 1, 168, false},
	{"Reserved", "RESERVE2", 3, 169, false},
	{"ZIP Code Tabulation Area (5-digit)", "ZCTA5", 5, 172, false},
	{"Subminor Civil Division", "SUBMCD", 5, 177, false},
	{"Subminor Civil Division Class Code", "SUBMCDCC", 2, 182, false},
	{"School District (Elementary)", "SDELM", 5, 184, false},
	{"School District (Secondary)", "SDSEC", 5, 189, false},
	{"School District (Unified)", "SDUNI", 5, 194, false},
	{"Area (Land)", "AREALAND", 14, 199, true},
	{"Area (Water)", "AREAWATR", 14, 213, true},
	{"Area Name-Legal/Statistical Area Description", "NAME", 90, 227, false},
	{"Functional Status Code", "FUNCSTAT", 1, 317, false},
	{"Geographic Change User Note Indicator", "GCUNI", 1, 318, false},
	{"Population Count (100%)", "POP100", 9, 319, true},
	{"Housing Unit Count (100%)", "HU100", 9, 328, true},
	{"Internal Point (Latitude)", "INTPTLAT", 11, 337, false},
	{"Internal Point (Longitude)", "INTPTLON", 12, 348, false},
	{"Legal/Statistical Area Description Code", "LSADC", 2, 360, false},
	{"Part Flag", "PARTFLAG", 1, 362, false},
	{"Reserved", "RESERVE3", 6, 363, false},
	{"Urban Growth Area", "UGA", 5, 369, false},
	{"State (ANSI)", "STATENS", 8, 374, false},
	{"County (ANSI)", "COUNTYNS", 8, 382, false},
	{"County Subdivision (ANSI)", "COUSUBNS", 8, 390, false},
	{"Place (ANSI)", "PLACENS", 8, 389, false},
	{"Consolidated City (ANSI)", "CONCITNS", 8, 406, false},
	{"American Indian Area/Alaska Native Area/Hawaiian Home Land (ANSI)", "AIANHHNS", 8, 414, false},
	{"American Indian Tribal Subdivision (ANSI)", "AITSNS", 8, 422, false},
	{"Alaska Native Regional Corporation (ANSI)", "ANRCNS", 8, 430, false},
	{"Subminor Civil Division (ANSI)", "SUBMCDNS", 8, 438, false},
	{"Congressional District (113th)", "CD113", 2, 446, false},
	{"Congressional District (114th)", "CD114", 2, 448, false},
	{"Congressional District (115th)", "CD115", 2, 450, false},
	{"State Legislative District (Upper Chamber) (Year 2)", "SLDU2", 3, 452, false},
	{"State Legislative District (Lower Chamber) (Year 2)", "SLDU3", 3, 455, false},
	{"State Legislative District (Upper Chamber) (Year 3)", "SLDU4", 3, 458, false},
	{"State Legislative District (Lower Chamber) (Year 3)", "SLDL2", 3, 461, false},
	{"State Legislative District (Upper Chamber) (Year 4)", "SLDL3", 3, 464, false},
	{"State Legislative District (Lower Chamber) (Year 4)", "SLDL4", 3, 476, false},
	{"American Indian Area/Alaska Native Area/Hawaiian Homeland size Code", "AIANHHSC", 2, 470, false},
	{"Combined Statistical Area Size Code", "CSASC", 2, 472, false},
	{"Combined NECTA Size Code", "CNECTASC", 2, 474, false},
	{"Metropolitan/Micropolitan Indicator", "MEMI", 1, 476, false},
	{"NECTA Metropolitan/Micropolitan Indicator", "NMEMI", 1, 477, false},
	{"Public Use Microdata Area", "PUMA", 5, 478, false},
	{"Reserved", "RESERVED", 18, 483, false},
}

func printUsage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", msg)
	}
	fmt.Fprintf(os.Stderr, "Usage: %s [census_data_folder]\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr,
		"('census_data_folder' defaults to the current working folder)\n",
	)
	os.Exit(1)
}

func openCensusDataFiles(censusDataFolder string) {
	log.Printf("Loading census data from %s", censusDataFolder)

	for _, fileName := range REQUIRED_FILES {
		filePath := path.Join(censusDataFolder, fileName)
		_, err := os.Stat(path.Join(censusDataFolder, fileName))
		if err != nil {
			printUsage(fmt.Sprintf(
				"Census data at %s is incomplete, missing file %s (%s)",
				censusDataFolder, fileName, err,
			))
		}

		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf(
				"Error opening census file %s (%s)\n", filePath, err,
			)
		}
		CENSUS_DATA_FILES[fileName] = CensusDataFile{
			File: file,
			Lock: &sync.Mutex{},
		}
	}
}

func oldReadFileLines(tableName string, file *os.File, lineChan chan string) {
	// buf := make([]byte, 64*1024)
	buf := make([]byte, 30)
	savedLine := make([]byte, 0, 64*1024)

	if _, err := file.Seek(0, 0); err != nil {
		log.Fatalf("Error rewinding file %s (%s)\n", file.Name(), err)
	}

	for {
		if bytesRead, err := file.Read(buf); err != nil {
			if bytesRead == 0 && err == io.EOF {
				break
			} else {
				log.Fatalf("Error reading from file %s (%s)\n", file.Name, err)
			}
		} else if bytesRead > 0 {
			lines := bytes.Split(buf, []byte{'\n'})
			lineCount := len(lines)
			if lineCount == 0 {
				continue
			}
			if lineCount >= 1 {
				savedLine = append(savedLine, lines[0]...)
			}
			if lineCount >= 2 {
				lineChan <- string(savedLine[:])
				savedLine = savedLine[0:0]
			}
			if lineCount > 2 {
				for _, line := range lines[1 : len(lines)-2] {
					lineChan <- string(line[:])
				}
			}
			if lineCount >= 2 {
				savedLine = append(savedLine, lines[len(lines)-1]...)
			}
		}
	}
	close(lineChan)
}

func openDB() {
	db, err := sql.Open(
		"postgres",
		"host=/var/run/postgresql dbname=census "+
			"user=census password=12106n13 sslmode=disable",
	)
	if err != nil {
		log.Fatalf(
			"Error connecting to database (%s)\n", err,
		)
	}
	DB = db
	DB.SetMaxOpenConns(MAX_DB_CONNECTIONS)
}

func closeDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection (%s)\n", err)
		}
	}
}

func reopenDB() {
	closeDB()
	openDB()
}

func GetAPIConcepts() map[string]APIConcept {
	log.Println("Downloading census API documentation")
	xmlAPIConcepts := new(XMLAPIConcepts)
	apiConcepts := make(map[string]APIConcept)

	resp, err := http.Get(DATA_DESCRIPTION_URL)
	if err != nil {
		log.Fatalln(err)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReader

	if err := decoder.Decode(xmlAPIConcepts); err != nil {
		log.Fatalln(err)
	}

	for _, xmlConcept := range xmlAPIConcepts.Concepts {
		var apiConcept APIConcept
		// apiConcept := new(APIConcept)

		match := CONCEPT_REGEXP.FindStringSubmatch(xmlConcept.Name)
		if len(match) == 0 {
			if xmlConcept.Name == "Geographic Characteristics" {
				apiConcept.Name = "geo_locations"
				apiConcept.Description = "Geographic Characteristics"
				apiConcept.VariableCount = 33
			} else if strings.HasPrefix(xmlConcept.Name, "PCT22A") {
				apiConcept.Name = "pct22a"
				apiConcept.Description =
					"GROUP QUARTERS POPULATION BY SEX BY GROUP QUARTERS " +
						"TYPE FOR THE POPULATION 18 YEARS AND OVER (WHITE ALONE)"
				apiConcept.VariableCount = 21
			} else if strings.HasPrefix(xmlConcept.Name, "PCT22D") {
				apiConcept.Name = "pct22d"
				apiConcept.Description =
					"GROUP QUARTERS POPULATION BY SEX BY GROUP QUARTERS " +
						"TYPE FOR THE POPULATION 18 YEARS AND OVER (ASIAN ALONE)"
				apiConcept.VariableCount = 21
			} else {
				log.Fatalf(
					"Concept %s does not match regex\n", xmlConcept.Name,
				)
			}
		} else {
			apiConcept.Name = strings.ToLower(match[1])
			apiConcept.Description = strings.TrimSpace(match[2])
			num, err := strconv.ParseInt(match[3], 10, 32)
			if err != nil {
				log.Fatalln(err)
			}
			apiConcept.VariableCount = int(num)
		}

		for _, variable := range xmlConcept.Variables {
			var apiVariable APIVariable

			apiVariable.Name = strings.ToLower(variable.Name)
			apiVariable.Description = variable.Description
			apiConcept.Variables = append(apiConcept.Variables, apiVariable)
		}
		if len(apiConcept.Variables) != apiConcept.VariableCount {
			log.Fatalf("Mismatched variable count for %s: %d != %d\n",
				apiConcept.Name,
				apiConcept.VariableCount,
				len(apiConcept.Variables),
			)
		}
		apiConcepts[apiConcept.Name] = apiConcept
	}

	return apiConcepts
}

func GetGeoLocations(queue chan *GeoLocation) {
	scanner := bufio.NewScanner(CENSUS_DATA_FILES[GEO_FILE].File)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		geoLocation := new(GeoLocation)
		for _, fd := range GeoLocationFieldDescriptions {
			var geoLocationField GeoLocationField

			startIndex := fd.Position - 1
			endIndex := startIndex + fd.Size
			if startIndex < 0 ||
				startIndex == endIndex ||
				startIndex >= len(line) ||
				endIndex > len(line) {
				log.Fatalf(
					"Value indices for %s (len: %d, %d:%d) out of range\n",
					fd.Name, len(line), startIndex, endIndex,
				)
			}
			geoLocationField.Name = fd.Name
			geoLocationField.ReferenceName = fd.ReferenceName
			geoLocationField.Size = fd.Size
			geoLocationField.Position = fd.Position
			geoLocationField.Numeric = fd.Numeric
			geoLocationField.Value = strings.TrimSpace(
				line[startIndex:endIndex],
			)
			geoLocation.Fields = append(geoLocation.Fields, geoLocationField)
		}
		queue <- geoLocation
	}
	close(queue)
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading packing list file (%s)", err)
	}
}

func GetDataTables(queue chan *CensusTable) {
	api := GetAPIConcepts()
	fileOffsets := make(map[string]int)

	scanner := bufio.NewScanner(CENSUS_DATA_FILES[PACKING_LIST_FILE].File)
	for scanner.Scan() {
		dataTable := new(CensusTable)

		line := scanner.Text()
		match := DATA_LOOKUP_REGEXP.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		tableName := match[1]
		num, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			log.Fatalf(
				"Error parsing file number from packing list file (%s)\n", err,
			)
		}
		fileNumber := int(num)
		num, err = strconv.ParseInt(match[3], 10, 32)
		if err != nil {
			log.Fatalf(
				"Error parsing column count from packing list file (%s)\n",
				err,
			)
		}
		columnCount := int(num)
		dataFileName := fmt.Sprintf(DATA_FILE_TEMPLATE, fileNumber)
		if _, present := CENSUS_DATA_FILES[dataFileName]; !present {
			log.Fatalf(
				"Census data file %s not recognized "+
					"(built from file number %d for table %s)\n",
				dataFileName, fileNumber, tableName,
			)
		}
		columnOffset := 0
		if offset, ok := fileOffsets[dataFileName]; ok {
			columnOffset = offset
		} else {
			fileOffsets[dataFileName] = 0
		}
		fileOffsets[dataFileName] += columnCount
		apiConcept, ok := api[tableName]
		if !ok {
			log.Fatalf("API data lacks table %s\n", tableName)
		}
		dataTable.DataLocation.DataFile = CENSUS_DATA_FILES[dataFileName]
		dataTable.DataLocation.ColumnOffset = columnOffset
		dataTable.DataLocation.ColumnCount = columnCount
		dataTable.Name = tableName
		dataTable.Description = apiConcept.Description
		dataTable.RowCount = 0
		dataTable.Columns = append(
			dataTable.Columns, CensusColumn{Name: "fileid"},
		)
		dataTable.Columns = append(
			dataTable.Columns, CensusColumn{Name: "stusab"},
		)
		dataTable.Columns = append(
			dataTable.Columns, CensusColumn{Name: "chariter"},
		)
		dataTable.Columns = append(
			dataTable.Columns, CensusColumn{Name: "cifsn"},
		)
		dataTable.Columns = append(
			dataTable.Columns, CensusColumn{Name: "logrecno"},
		)
		for _, apiVariable := range apiConcept.Variables {
			dataTable.Columns = append(
				dataTable.Columns, CensusColumn{Name: apiVariable.Name},
			)
		}

		if len(dataTable.Columns)-5 != dataTable.DataLocation.ColumnCount {
			log.Fatalf("Found column count mismatch in table %s (%d != %d)\n",
				dataTable.Name, len(dataTable.Columns),
				dataTable.DataLocation.ColumnCount,
			)
		}
		queue <- dataTable
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading packing list file (%s)", err)
	}
	close(queue)
}

func loadGeoLocationData(geoLocationDataLoaded chan bool) {
	log.Println("Loading geographic location data")

	dropGeoLocationsTableQuery := "DROP TABLE geo_locations"
	createGeoLocationsTableQuery :=
		"CREATE TABLE geo_locations (id SERIAL PRIMARY KEY"
	for _, gvfd := range GeoLocationFieldDescriptions {
		if gvfd.ReferenceName == "AREALAND" ||
			gvfd.ReferenceName == "AREAWATR" {
			createGeoLocationsTableQuery += fmt.Sprintf(
				", %s bigint", gvfd.ReferenceName,
			)
		} else if gvfd.Numeric {
			createGeoLocationsTableQuery += fmt.Sprintf(
				", %s integer", gvfd.ReferenceName,
			)
		} else {
			createGeoLocationsTableQuery += fmt.Sprintf(
				", %s varchar (%d)", gvfd.ReferenceName, gvfd.Size,
			)
		}
	}
	createGeoLocationsTableQuery += ")"

	if tx, txErr := DB.Begin(); txErr == nil {
		if _, dtErr := tx.Exec(dropGeoLocationsTableQuery); dtErr == nil {
			if _, ctErr := tx.Exec(createGeoLocationsTableQuery); ctErr == nil {
				if err := tx.Commit(); err == nil {
					log.Println("Created table 'geo_locations'")
				} else {
					log.Fatalf("Error committing transaction (%s)\n", err)
				}
			} else if rbErr := tx.Rollback(); rbErr == nil {
				log.Fatalf("Error creating table 'geo_locations'\n"+
					"Query error: %s\n"+
					"Query: %s\n",
					ctErr, createGeoLocationsTableQuery,
				)
			} else {
				log.Fatalf("Error rolling back transaction\n"+
					"Query error: %s\n"+
					"Rollback error: %s\n"+
					"Query: %s\n",
					ctErr, rbErr, createGeoLocationsTableQuery,
				)
			}
		} else if rbErr := tx.Rollback(); rbErr == nil {
			log.Printf("Error dropping table 'geo_locations'\n"+
				"Query error: %s\n"+
				"Query: %s\n",
				dtErr, dropGeoLocationsTableQuery,
			)
		} else {
			log.Fatalf("Error rolling back transaction\n"+
				"Query error: %s\n"+
				"Rollback error: %s\n"+
				"Query: %s\n",
				dtErr, rbErr, dropGeoLocationsTableQuery,
			)
		}
	} else {
		log.Fatalf("Error beginning transaction (%s)\n", txErr)
	}

	geoTableColumnNameSlice := make([]string, len(GeoLocationFieldDescriptions))
	for fi, fd := range GeoLocationFieldDescriptions {
		geoTableColumnNameSlice[fi] = fd.ReferenceName
	}
	geoTableColumnNames := strings.Join(geoTableColumnNameSlice, ", ")

	geoLocationQueue := make(chan *GeoLocation, 10)
	go GetGeoLocations(geoLocationQueue)
	tx, err := DB.Begin()
	if err != nil {
		log.Fatalf("Error beginning transaction (%s)\n", err)
	}
	for geoLocation := range geoLocationQueue {
		rowValueSlice := make([]string, len(geoLocation.Fields))
		for fi, field := range geoLocation.Fields {
			if field.Numeric {
				rowValueSlice[fi] = field.Value
			} else {
				rowValueSlice[fi] = fmt.Sprintf("'%s'", field.Value)
			}
		}
		geoLocationQuery := fmt.Sprintf(
			"INSERT INTO geo_locations (%s) VALUES (%s)",
			geoTableColumnNames, strings.Join(rowValueSlice, ", "),
		)
		if _, err := tx.Exec(geoLocationQuery); err != nil {
			log.Fatalf("Error INSERTing geographic location\n"+
				"Query error: %s\n"+
				"Query: %s\n",
				err, geoLocationQuery,
			)
		} else if PRINT_SQL_QUERIES {
			log.Println(geoLocationQuery)
		}
	}
	if err = tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction (%s)\n", err)
	}

	geoLocationDataLoaded <- true
}

func loadCensusData(censusDataLoaded chan bool) {
	dataTableQueue := make(chan *CensusTable, 10)
	tableLoadedChan := make(chan string, 10)

	go GetDataTables(dataTableQueue)

	tableCount := 0
	for dataTable := range dataTableQueue {
		log.Printf("Loading census data table %s\n", dataTable.Name)
		go loadCensusDataTable(dataTable, tableLoadedChan)
		tableCount++
	}

	for {
		loadedTableName := <-tableLoadedChan
		tableCount--
		log.Printf("Loaded table %s, %d to go\n", loadedTableName, tableCount)
		if tableCount == 0 {
			break
		}
	}

	censusDataLoaded <- true
}

func loadCensusDataTable(dataTable *CensusTable, tableLoaded chan string) {
	startIndex := dataTable.DataLocation.ColumnOffset + 5
	endIndex := startIndex + dataTable.DataLocation.ColumnCount

	columnDefinitionSlice := make([]string, len(dataTable.Columns)-5)
	for ci, column := range dataTable.Columns {
		if ci < 5 { // skip the first 5 geographic location columns
			continue
		}
		columnDefinitionSlice[ci-5] = fmt.Sprintf(
			"%s integer", column.Name,
		)
	}
	columnDefinitions := strings.Join(columnDefinitionSlice, ", ")
	dropTableQuery := fmt.Sprintf("DROP TABLE %s", dataTable.Name)
	createTableQuery := fmt.Sprintf(
		"CREATE TABLE %s ("+
			"id SERIAL PRIMARY KEY, fileid varchar(6), stusab varchar(2), "+
			"chariter varchar(3), cifsn varchar(3), logrecno varchar(7), %s"+
			")", dataTable.Name, columnDefinitions,
	)

	if tx, txErr := DB.Begin(); txErr == nil {
		if _, dtErr := tx.Exec(dropTableQuery); dtErr == nil {
			if _, ctErr := tx.Exec(createTableQuery); ctErr == nil {
				if err := tx.Commit(); err == nil {
					log.Printf("Created table '%s'\n", dataTable.Name)
				} else {
					log.Fatalf("Error committing transaction (%s)\n", err)
				}
			} else if rbErr := tx.Rollback(); rbErr == nil {
				log.Fatalf("Error creating table '%s'\n"+
					"Query error: %s\n"+
					"Query: %s\n",
					dataTable.Name, ctErr, createTableQuery,
				)
			} else {
				log.Fatalf("Error rolling back transaction\n"+
					"Query error: %s\n"+
					"Rollback error: %s\n"+
					"Query: %s\n",
					ctErr, rbErr, createTableQuery,
				)
			}
		} else if rbErr := tx.Rollback(); rbErr == nil {
			log.Printf("Error dropping table '%s'\n"+
				"Query error: %s\n"+
				"Query: %s\n",
				dataTable.Name, dtErr, dropTableQuery,
			)
		} else {
			log.Fatalf("Error rolling back transaction\n"+
				"Query error: %s\n"+
				"Rollback error: %s\n"+
				"Query: %s\n",
				dtErr, rbErr, dropTableQuery,
			)
		}
	} else {
		log.Fatalf("Error beginning transaction (%s)\n", txErr)
	}

	columnNameSlice := make([]string, len(dataTable.Columns))
	for ci, column := range dataTable.Columns {
		columnNameSlice[ci] = column.Name
	}
	columnNames := strings.Join(columnNameSlice, ", ")

	dataTable.DataLocation.DataFile.Lock.Lock()
	defer dataTable.DataLocation.DataFile.Lock.Unlock()

	lineCount := 0
	rowCount := 0
	if _, err := dataTable.DataLocation.DataFile.File.Seek(0, 0); err != nil {
		log.Fatalf("Error rewinding file %s (%s)\n",
			dataTable.DataLocation.DataFile.File.Name(),
			err,
		)
	}
	scanner := bufio.NewScanner(dataTable.DataLocation.DataFile.File)
	tx, err := DB.Begin()
	if err != nil {
		log.Fatalf("Error beginning transaction (%s)\n", err)
	}
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		rowValues := strings.Split(line, ",")
		if startIndex < 5 || endIndex > len(rowValues) {
			log.Fatalf(
				"Row value indices for %s out of range: [%d:%d] (%d) =>\n"+
					"rowValues: %s\n"+
					"file: %s\n"+
					"line (%d, %d): %s\n",
				dataTable.Name, startIndex, endIndex, len(rowValues),
				rowValues, dataTable.DataLocation.DataFile.File.Name(),
				lineCount, len(line), line,
			)
		}
		tableValues := append(rowValues[0:5], rowValues[startIndex:endIndex]...)
		for ci, columnValue := range tableValues {
			dataTable.Columns[ci].Value = columnValue
		}
		columnValueSlice := make([]string, len(dataTable.Columns))
		for ci, column := range dataTable.Columns {
			if ci < 5 {
				columnValueSlice[ci] = fmt.Sprintf("'%s'", column.Value)
			} else {
				columnValueSlice[ci] = column.Value
			}
		}
		dataQuery := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			dataTable.Name, columnNames, strings.Join(columnValueSlice, ", "),
		)
		if _, err := tx.Exec(dataQuery); err != nil {
			log.Fatalf("Error INSERTing census data into %s\n"+
				"Query error: %s\n"+
				"Query: %s\n",
				dataTable.Name, err, dataQuery,
			)
		} else if PRINT_SQL_QUERIES {
			log.Println(dataQuery)
		}
		rowCount++
	}
	if err = tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction (%s)\n", err)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from file %s (%s)\n",
			dataTable.DataLocation.DataFile.File.Name(), err,
		)
	}
	log.Printf("%s: wrote %d rows\n", dataTable.Name, rowCount)

	tableLoaded <- dataTable.Name
}

func main() {
	var censusDataFolder string
	geoLocationDataLoaded := make(chan bool)
	censusDataLoaded := make(chan bool)

	// Check for a specified census data folder
	if len(os.Args) == 1 {
		dataFolder, err := os.Getwd()
		if err != nil {
			log.Fatalf(
				"Could not get current working folder: %s\n", err,
			)
		}
		censusDataFolder = dataFolder
	} else if len(os.Args) == 2 {
		censusDataFolder = os.Args[1]
	} else {
		printUsage("")
	}

	// Check that the census data folder exists and is a folder
	fi, err := os.Stat(censusDataFolder)
	if err != nil {
		printUsage(fmt.Sprintf(
			"Could not locate folder %s", censusDataFolder,
		))
	}
	if !fi.IsDir() {
		printUsage(fmt.Sprintf(
			"Folder %s is not a folder", censusDataFolder,
		))
	}

	openCensusDataFiles(censusDataFolder)
	openDB()

	// Spawn goroutines to load the data
	go loadGeoLocationData(geoLocationDataLoaded)
	go loadCensusData(censusDataLoaded)

	// Wait for the goroutines to finish
	<-censusDataLoaded
	log.Println("Census data loaded")
	<-geoLocationDataLoaded
	log.Println("Geographic location data loaded")

	// Done!
	log.Println("Loading complete")
}
