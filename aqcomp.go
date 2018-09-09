package main

import (
	"bitbucket.org/ctessum/cdf"
	"encoding/csv"
	"fmt"
	"github.com/fatih/structs"
	//	"io/ioutil" <-- ioutil.ReadDir returns an interface instead of
	//	[]string, so use path/filepath instead.
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// listFiles will list all the csv files for which there is measurement
// data from OpenAQ.
func listFiles(csvFolder string) []string {
	var csvList []string

	err := filepath.Walk(csvFolder, func(path string, info os.FileInfo, err error) error {
		csvList = append(csvList, path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	//	files, err := ioutil.ReadDir(csvFolder)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	for _, f := range files {
	//		csvList = append(csvList, f)
	//	}

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
	GEOSlat float64
	GEOSlon float64
	// A field for the corresponding GEOS-Chem simulation time
	GEOStime string
	GEOShour int
	//	The PM2.5 chem value calculated from the GEOS-Chem simulation
	PM25  float32
	NH4   float32
	NIT   float32
	SO4   float32
	BCPI  float32
	BCPO  float32
	OCPI  float32
	OCPO  float32
	DST1  float32
	DST2  float32
	SALA  float32
	TSOA0 float32
	TSOA1 float32
	TSOA2 float32
	TSOA3 float32
	ISOA1 float32
	ISOA2 float32
	ISOA3 float32
	ASOAN float32
	ASOA1 float32
	ASOA2 float32
	ASOA3 float32
}

// lats are the grid cell latitudes. They should be ordered from
// smallest to largest.
var lats = []float64{-89.5, -88, -86, -84, -82, -80, -78, -76, -74, -72, -70, -68, -66, -64, -62, -60, -58, -56, -54, -52, -50, -48, -46, -44, -42, -40, -38, -36, -34, -32, -30, -28, -26, -24, -22, -20, -18, -16, -14, -12, -10, -8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 89.5}

// lons are the grid cell longitudes. They should be ordered from
// smallest to largest.
var lons = []float64{-180, -177.5, -175, -172.5, -170, -167.5, -165, -162.5, -160, -157.5, -155, -152.5, -150, -147.5, -145, -142.5, -140, -137.5, -135, -132.5, -130, -127.5, -125, -122.5, -120, -117.5, -115, -112.5, -110, -107.5, -105, -102.5, -100, -97.5, -95, -92.5, -90, -87.5, -85, -82.5, -80, -77.5, -75, -72.5, -70, -67.5, -65, -62.5, -60, -57.5, -55, -52.5, -50, -47.5, -45, -42.5, -40, -37.5, -35, -32.5, -30, -27.5, -25, -22.5, -20, -17.5, -15, -12.5, -10, -7.5, -5, -2.5, 0, 2.5, 5, 7.5, 10, 12.5, 15, 17.5, 20, 22.5, 25, 27.5, 30, 32.5, 35, 37.5, 40, 42.5, 45, 47.5, 50, 52.5, 55, 57.5, 60, 62.5, 65, 67.5, 70, 72.5, 75, 77.5, 80, 82.5, 85, 87.5, 90, 92.5, 95, 97.5, 100, 102.5, 105, 107.5, 110, 112.5, 115, 117.5, 120, 122.5, 125, 127.5, 130, 132.5, 135, 137.5, 140, 142.5, 145, 147.5, 150, 152.5, 155, 157.5, 160, 162.5, 165, 167.5, 170, 172.5, 175, 177.5}

// findLatLon finds the latitude or longitude of the GEOS-Chem
// simulation grid cell corresponding to the latitude or longitude
// of the measurement (given as a string).
func findLatLon(measuredLat string, lat []float64) float64 {
	f, err := strconv.ParseFloat(measuredLat, 64)
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

	f, err := strconv.Atoi(measuredHour)
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

		for _, line := range lines {
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
		}
		f.Close()
	}

}

// writeMeasurements will:
// open the NetCDF file associated with each GEOStimestring,
// find the correct measurement, and add that to the struct too;
// write the structs to a csv file called "output.csv".
// writeMeasurements takes in the hour, latitude and longitude as
// integers corresponding to their location in the file. The vertical
// levels are not inputs because we assume the levels corresponding to
// the measurements to be at the surface (i.e., 0).
func writeMeasurements(ms Measurements, chemFolder string) {

	ff, _ := os.Open(chemFolder + ms.GEOStime)
	defer ff.Close()
	f, _ := cdf.Open(ff)

	ms.ASOA1 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ASOA1")
	ms.ASOA2 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ASOA2")
	ms.ASOA3 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ASOA3")
	ms.ASOAN = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ASOAN")
	ms.ISOA1 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ISOA1")
	ms.ISOA2 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ISOA2")
	ms.ISOA3 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__ISOA3")
	ms.TSOA0 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__TSOA0")
	ms.TSOA1 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__TSOA1")
	ms.TSOA2 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__TSOA2")
	ms.TSOA3 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__TSOA3")
	ms.DST1 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__DST1")
	ms.DST2 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__DST2")
	ms.SALA = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__SALA")
	ms.OCPI = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__OCPI")
	ms.OCPO = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__OCPO")
	ms.BCPI = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__BCPI")
	ms.BCPO = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__BCPO")
	ms.SO4 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__SO4")
	ms.NIT = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__NIT")
	ms.NH4 = varReading(ms.GEOShour, ms.GEOSlat, ms.GEOSlon, f, "IJ_AVG_S__NH4")

	ms.PM25 = 1.33*(ms.NH4+ms.NIT+ms.SO4) + ms.BCPI + ms.BCPO + 2.1*(ms.OCPO+1.16*ms.OCPI) + ms.DST1 + 0.38*ms.DST2 + 1.86*ms.SALA + 1.16*(ms.TSOA0+ms.TSOA1+ms.TSOA2+ms.TSOA3+ms.ISOA1+ms.ISOA2+ms.ISOA3+ms.ASOAN+ms.ASOA1+ms.ASOA2+ms.ASOA3)

}

func varReading(hour int, lat, lon float64, f *cdf.File, pol string) float32 {

	lev := 0.0
	indexx := int(lon + lat*47 + lev*144)

	dims := f.Header.Lengths(pol)
	if len(dims) == 0 {
		panic(fmt.Errorf("%v isn't on file", pol))
	}
	dims = dims[1:]
	// This is done because the 0th entry in dims is 0.
	nread := 1
	for _, dim := range dims {
		nread *= dim
	}

	start, end := make([]int, len(dims)+1), make([]int, len(dims)+1)
	start[0], end[0] = hour, hour+1

	r := f.Reader(pol, start, end)
	buf := r.Zero(nread)
	_, err := r.Read(buf)
	if err != nil {
		panic(err)
	}
	// The following ought to be passed to ms.
	return buf.([]float32)[indexx]

}

func csvWriter(ms Measurements) {
	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}

	writefile := csv.NewWriter(file)
	defer writefile.Flush()

	for _, v := range structs.Values(ms) {
		err := writefile.Write(v.([]string))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	csvFolder := "/home/marshall/sthakrar/2015openaqdata/csvfiles/"
	cm := make(chan Measurements, 1000)
	go readMeasurements(csvFolder)
	go writeMeasurements(<-cm, "/home/marshall/sthakrar/go/src/github.com/spatialmodel/inmap/cmd/inmap/testdata/preproc/GlobalTestData")
	// etc.
}
