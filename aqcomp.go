package main

import (
	"bitbucket.org/ctessum/cdf"
	"encoding/csv"
	"fmt"
	"github.com/fatih/structs"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// listFiles will list all the csv files for which there is measurement
// data from OpenAQ.
func listFiles(csvFolder string) []string {
	var csvList []string

	files, err := ioutil.ReadDir(csvFolder)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		csvList = append(csvList, f)
	}
	return csvList
}

// Measurements has measurement data in standard units, information on
// the GEOS-Chem grid cell at which the measurement takes place,
// the time of the measurement, and the corresponding GEOS-Chem
// simulation value.
type Measurements struct {
	time      string
	pollutant string
	value     string
	unit      string
	latitude  string
	longitude string
	// Fields for the corresponding GEOS-Chem grid cell
	GEOSlat float32
	GEOSlon float32
	// A field for the corresponding GEOS-Chem simulation time
	GEOStime string
	// A field for the simulation data
	chemValue string
}

// lats are the grid cell latitudes. They should be ordered from
// smallest to largest.
var lats = []float32{-89.5, -88, -86, -84, -82, -80, -78, -76, -74, -72, -70, -68, -66, -64, -62, -60, -58, -56, -54, -52, -50, -48, -46, -44, -42, -40, -38, -36, -34, -32, -30, -28, -26, -24, -22, -20, -18, -16, -14, -12, -10, -8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 89.5}

// lons are the grid cell longitudes. They should be ordered from
// smallest to largest.
var lons = []float32{-180, -177.5, -175, -172.5, -170, -167.5, -165, -162.5, -160, -157.5, -155, -152.5, -150, -147.5, -145, -142.5, -140, -137.5, -135, -132.5, -130, -127.5, -125, -122.5, -120, -117.5, -115, -112.5, -110, -107.5, -105, -102.5, -100, -97.5, -95, -92.5, -90, -87.5, -85, -82.5, -80, -77.5, -75, -72.5, -70, -67.5, -65, -62.5, -60, -57.5, -55, -52.5, -50, -47.5, -45, -42.5, -40, -37.5, -35, -32.5, -30, -27.5, -25, -22.5, -20, -17.5, -15, -12.5, -10, -7.5, -5, -2.5, 0, 2.5, 5, 7.5, 10, 12.5, 15, 17.5, 20, 22.5, 25, 27.5, 30, 32.5, 35, 37.5, 40, 42.5, 45, 47.5, 50, 52.5, 55, 57.5, 60, 62.5, 65, 67.5, 70, 72.5, 75, 77.5, 80, 82.5, 85, 87.5, 90, 92.5, 95, 97.5, 100, 102.5, 105, 107.5, 110, 112.5, 115, 117.5, 120, 122.5, 125, 127.5, 130, 132.5, 135, 137.5, 140, 142.5, 145, 147.5, 150, 152.5, 155, 157.5, 160, 162.5, 165, 167.5, 170, 172.5, 175, 177.5}

// findLatLon finds the latitude or longitude of the GEOS-Chem
// simulation grid cell corresponding to the latitude or longitude
// of the measurement (given as a string).
func findLatLon(measuredLat string, lat []float32) float32 {
	f, err = strconv.ParseFloat(measuredLat, 32)
	if err != nil {
		panic(err)
	}
	i := len(lat) - 1
	for f < lat[i] {
		i -= 1
	}
	return lat[i]
}

// We also want the measurement time. The GEOS-Chem time
// variable is in hours since 1985-1-1 00:00:0.0 (including
// 7 leap days). There are eight records a day, one every
// three hours, starting from 03:00 (2015-01-01 at 00:00 is
// 262968 hours since the start time). Z time is reported
// for both simulation and measurement, I believe; however,
// the averaging periods of the measurements may differ.
// For now, I will simply parse the measurement time and
// assign it to n int, where the nth record in each day of
// GEOS-Chem simulation corresponds to the time interval.
func findTime(measuredHour string) int {

	f, err = strconv.Atoi(measuredHour)
	if err != nil {
		panic(err)
	}
	switch {
	case f <= 3:
		return 1
	case f <= 6:
		return 2
	case f <= 9:
		return 3
	case f <= 12:
		return 4
	case f <= 15:
		return 5
	case f <= 18:
		return 6
	case f <= 21:
		return 7
	case f <= 24:
		return 8
	}
	return 0
}

// readMeasurements reads the measurement PM2.5 data (value, lat, lon,
// time, and units) from all csv files in the folder, and returns the
// values in standard units, the GEOS-Chem grid cell information, and
// the time.
func readMeasurements(csvFolder string) {

	// List the files in the folder.
	csvList := listFiles(csvFolder)

	// Declare a buffered channel for the measurements to go through.
	var cm chan Measurements

	// Open all the files and pass each measurement to the buffered
	// channel.
	for _, filename := range csvList {

		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		lines, err := csv.NewReader(f).ReadAll()
		if err != nil {
			panic(err)
		}

		for i, line := range lines {
			// for each Measurement, read the time
			// field, which is in the standard form
			// [YYYY]-[MM]-[DD]T[HH]:[MM]:[SS].000Z.
			// you want to parse this to make a string
			// for the NetCDF filename, which is in
			// the format "ts.[YYYY][MM][DD].000000.nc".
			GEOStimestring := "ts." + line[3][:4] + line[3][6:7] + line[3][9:10] + ".000000.nc"

			// Pass the measurements from each csv file
			// to the buffered channel.
			cm <- Measurements{
				time:      line[3],
				pollutant: line[5],
				value:     line[6],
				unit:      line[7],
				latitude:  line[8],
				longitude: line[9],
				GEOStime:  GEOStimestring,
				GEOSlat:   findLatLon(line[8], lats),
				GEOSlon:   findLatLon(line[9], lons),
				GEOShour:  findTime(line[3][12:13]),
			}
			fmt.Println(data.latitude + " " + data.longitude + " " + data.value)
		}
		f.Close()
	}

}

// writeMeasurements will:
// open the NetCDF file associated with each GEOStimestring,
// find the correct measurement, and add that to the struct too;
// write the structs to a csv file called "output.csv".
func writeMeasurements(ms Measurements, csvFolder string) {

	//	ms.pollutant should determine the string, but for now we're only
	//	concerned with PM2.5, which requires reading NH4, NIT, SO4,
	//	BCPI, BCPO, OCPI, OCPO, DST1, DST2, SALA, TSOA0, TSOA1, TSOA2,
	//	TSOA3, ISOA1, ISOA2, ISOA3, ASOAN [which I forgot to write out],
	//	ASOA1, ASOA2, and ASOA3.

	ff, _ := os.Open(csvFolder + ms.GEOStimestring)
	f, _ := cdf.Open(ff)
	defer f.Close()
	r = f.Reader(pol)
	// I haven't done this yet
}

/*
	time      string
	pollutant string
	value     string
	unit      string
	latitude  string
	longitude string
	GEOSlat float32
	GEOSlon float32
	GEOStime string
	chemValue string
*/

func csvWriter(ms Measurements) {
	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}

	writefile := csv.NewWriter(file)
	defer writer.Flush()

	for _, v := range structs.Values(ms) {
		err := writer.Write(v.(string))
	}
}

//csvFiles
//"/home/marshall/sthakrar/2015openaqdata/csvfiles/[DATE].csv",

func main() {
	csvFolder := "/home/marshall/sthakrar/2015openaqdata/csvfiles/"
	cm := make(chan Measurements, 1000)
	go readMeasurements(csvFolder)
	go writeMeasurements(<-cm, csvFolder)
	// etc.
}
