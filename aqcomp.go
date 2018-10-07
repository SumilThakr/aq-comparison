package main

import (
	"bitbucket.org/ctessum/cdf"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const STP_P = 1013.25
const STP_T = 298.
const ppb_ugm3 = (1000000.0 / 8.314) * 100.0 * STP_P / (STP_T * 1000000000.0)

var MWaer = [8]float32{18, 12, 12, 62, 96, 29, 31.4, 150} // NH4, EC, OC, NIT, SO4, DUST, SALA, SOA

type XY = []string

//	sPM string
//	mPM string

type outputComp struct {
	measuredPM  string
	simulatedPM string
	time        string
	GEOShour    int
	lat         int
	lon         int
}

type ms struct {
	csvPath string
	ncfPath string
	date    time.Time
	results []outputComp
}

func varReading(hour, lat, lon int, f *cdf.File, pol string) float32 {

	lev := 0
	// This can't be right:
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
	// fmt.Println(indexx)
	return buf.([]float32)[indexx]

}

func listFiles(csvFolder string) ([]string, error) {
	var csvList []string

	err := filepath.Walk(csvFolder, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".csv") {
			csvList = append(csvList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	//  N.B. The first entry in csvList would have been the directory itself, but
	//  this was taken out by the if strings.HasSuffix statement. If there
	//  are no elements in the list, return an error.
	if len(csvList) == 0 {
		return nil, fmt.Errorf("No csv files in the measurement folder: %v", csvFolder)
	}

	return csvList, nil
}

func initMs(csvFolder string, ncfFolder string) []ms {

	if !strings.HasSuffix(ncfFolder, "/") {
		ncfFolder = ncfFolder + "/"
	}

	csvList, err := listFiles(csvFolder)
	if err != nil {
		log.Fatal(err)
	}

	var sliceMs []ms

	for _, file := range csvList {

		dateHyphen := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		t, err := time.Parse("2006-01-02", dateHyphen)
		if err != nil {
			fmt.Println(err)
		}

		ncfFormat := ncfFolder + "ts." + t.Format("20060102.150405") + ".nc"

		newMs := ms{
			csvPath: file,
			date:    t,
			ncfPath: ncfFormat,
		}
		sliceMs = append(sliceMs, newMs)

	}
	return sliceMs
}

// *************************************************************************
// *************************************************************************
//                          READING THE CSV FILE
// *************************************************************************
// *************************************************************************

func initResults(mh ms) ([]outputComp, error) {
	var outputResults []outputComp

	ff, err := os.Open(mh.ncfPath)
	if err != nil {
		return nil, fmt.Errorf("%s cannot be opened: %v", mh.ncfPath, err)
	}
	defer ff.Close()
	f, err := cdf.Open(ff)
	if err != nil {
		return nil, fmt.Errorf("%s cannot be opened: %v", mh.ncfPath, err)
	}

	csvf, errOpen := os.Open(mh.csvPath)
	if errOpen != nil {
		return nil, fmt.Errorf("The csv %s cannot be opened: %v", mh.csvPath, errOpen)
	}
	r := csv.NewReader(csvf)
	lines, errRead := r.ReadAll()
	if errRead != nil {
		return nil, fmt.Errorf("reading %s isn't working: %v", mh.csvPath, errRead)
	}
	//  We remove the header information. Ideally, we would like to use this
	//  information to allow for differently structured data.
	lines = lines[1:]
	//  For each line, we want to save out the time, GEOStime, lat and lon.
	//  But, we only want to select those that are PM2.5 measurements, for
	//  now.
	for _, line := range lines {
		foundTime, errTime := findTime(line[3][12:13])
		foundLat, errLat := findLatLon(line[8], lats)
		foundLon, errLon := findLatLon(line[9], lons)

		if line[5] == "pm25" && errTime == nil && errLat == nil && errLon == nil {

			ASOA1 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ASOA1")
			ASOA2 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ASOA2")
			ASOA3 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ASOA3")
			//ASOAN := ppb_ugm3*MWaer[7]*varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ASOAN")
			ISOA1 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ISOA1")
			ISOA2 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ISOA2")
			ISOA3 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__ISOA3")
			TSOA0 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__TSOA0")
			TSOA1 := ppb_ugm3 * MWaer[7] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__TSOA1")
			TSOA2 := ppb_ugm3 * MWaer[6] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__TSOA2")
			TSOA3 := ppb_ugm3 * MWaer[6] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__TSOA3")
			DST1 := ppb_ugm3 * MWaer[5] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__DST1")
			DST2 := ppb_ugm3 * MWaer[5] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__DST2")
			SALA := ppb_ugm3 * MWaer[6] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__SALA")
			OCPI := ppb_ugm3 * MWaer[2] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__OCPI")
			OCPO := ppb_ugm3 * MWaer[2] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__OCPO")
			BCPI := ppb_ugm3 * MWaer[1] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__BCPI")
			BCPO := ppb_ugm3 * MWaer[1] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__BCPO")
			SO4 := ppb_ugm3 * MWaer[4] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__SO4")
			NIT := ppb_ugm3 * MWaer[3] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__NIT")
			NH4 := ppb_ugm3 * MWaer[0] * varReading(foundTime, foundLat, foundLon, f, "IJ_AVG_S__NH4")

			// Below is the correct ms.PM25 value. However, I forgot to write
			// out ASOAN. So I have commented this out and added a new PM2.5
			// value without ASOAN for now.
			//ms.PM25 = 1.33*(ms.NH4+ms.NIT+ms.SO4) + ms.BCPI + ms.BCPO +
			//2.1*(ms.OCPO+1.16*ms.OCPI) + ms.DST1 + 0.38*ms.DST2 + 1.86*ms.SALA + 1.16*(ms.TSOA0+ms.TSOA1+ms.TSOA2+ms.TSOA3+ms.ISOA1+ms.ISOA2+ms.ISOA3+ms.ASOAN+ms.ASOA1+ms.ASOA2+ms.ASOA3)
			simPM := 1.33*(NH4+NIT+SO4) + BCPI + BCPO + 2.1*(OCPO+1.16*OCPI) + DST1 + 0.38*DST2 + 1.86*SALA + 1.16*(TSOA0+TSOA1+TSOA2+TSOA3+ISOA1+ISOA2+ISOA3+ASOA1+ASOA2+ASOA3)
			simPM = simPM * 150 / 28.97

			result := outputComp{
				time:        line[3],
				measuredPM:  line[6],
				GEOShour:    foundTime,
				lat:         foundLat,
				lon:         foundLon,
				simulatedPM: fmt.Sprintf("%f", simPM),
			}
			outputResults = append(outputResults, result)
		}
	}
	return outputResults, nil
}

func findTime(measuredHour string) (int, error) {
	f, err := strconv.Atoi(measuredHour)
	if err != nil {
		return 0, err
	}
	switch {
	case f >= 0 && f <= 3:
		return 1, nil
	case f <= 6:
		return 2, nil
	case f <= 9:
		return 3, nil
	case f <= 12:
		return 4, nil
	case f <= 15:
		return 5, nil
	case f <= 18:
		return 6, nil
	case f <= 21:
		return 7, nil
	case f <= 24:
		return 8, nil
	}
	return 0, fmt.Errorf("%d is not an integer between 0 and 24", f)
}
func findLatLon(measuredLat string, lat []float64) (int, error) {

	f, err := strconv.ParseFloat(measuredLat, 64)
	if err != nil {
		return 0, fmt.Errorf("The lat/lon is either absent or not parseable: %s", measuredLat)
	}

	if math.Abs(f) > lat[len(lat)-1] {
		return 0, fmt.Errorf("the latitude or longitude is out of bounds: %s", measuredLat)
	}
	i := len(lat) - 1
	for f < lat[i] {
		i -= 1
	}
	return i, nil
}

// lats and lons are the grid cell latitudes and longitudes,
// respectively. They should be ordered from smallest to largest. These
// correspond to GEOS-Chem simulations at 2x2.5. Ideally, they should be
// selected by the user depending on the run specification.
var lats = []float64{-89.5, -88, -86, -84, -82, -80, -78, -76, -74, -72, -70, -68, -66, -64, -62, -60, -58, -56, -54, -52, -50, -48, -46, -44, -42, -40, -38, -36, -34, -32, -30, -28, -26, -24, -22, -20, -18, -16, -14, -12, -10, -8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 89.5}

var lons = []float64{-180, -177.5, -175, -172.5, -170, -167.5, -165, -162.5, -160, -157.5, -155, -152.5, -150, -147.5, -145, -142.5, -140, -137.5, -135, -132.5, -130, -127.5, -125, -122.5, -120, -117.5, -115, -112.5, -110, -107.5, -105, -102.5, -100, -97.5, -95, -92.5, -90, -87.5, -85, -82.5, -80, -77.5, -75, -72.5, -70, -67.5, -65, -62.5, -60, -57.5, -55, -52.5, -50, -47.5, -45, -42.5, -40, -37.5, -35, -32.5, -30, -27.5, -25, -22.5, -20, -17.5, -15, -12.5, -10, -7.5, -5, -2.5, 0, 2.5, 5, 7.5, 10, 12.5, 15, 17.5, 20, 22.5, 25, 27.5, 30, 32.5, 35, 37.5, 40, 42.5, 45, 47.5, 50, 52.5, 55, 57.5, 60, 62.5, 65, 67.5, 70, 72.5, 75, 77.5, 80, 82.5, 85, 87.5, 90, 92.5, 95, 97.5, 100, 102.5, 105, 107.5, 110, 112.5, 115, 117.5, 120, 122.5, 125, 127.5, 130, 132.5, 135, 137.5, 140, 142.5, 145, 147.5, 150, 152.5, 155, 157.5, 160, 162.5, 165, 167.5, 170, 172.5, 175, 177.5}

func csvWriter(filename string, tWrt []XY) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Cannot write to file: %v", err)
	}

	writefile := csv.NewWriter(file)

	for _, xys := range tWrt {
		if err2 := writefile.Write(xys); err2 != nil {
			return fmt.Errorf("Cannot write to file: %v", err2)
		}
	}

	writefile.Flush()
	return nil
}

// *************************************************************************
// *************************************************************************
//                          MAIN FUNCTION BELOW
// *************************************************************************
// *************************************************************************

func main() {

	// For some folder, list the files. The definition of the folder, as
	// of now, ends with a backslash. Ideally, this would allow for this to
	// be missing.
	csvFolder := "/home/marshall/sthakrar/2015openaqdata/csvfiles"
	// Test folder is the following:
	//csvFolder := "/home/marshall/sthakrar/go/src/github.com/SumilThakr/aqcomp/testfiles/threedays/"
	ncfFolder := "/home/hill0408/sthakrar/Runs/globnosoan/"
	outputFolder := "/home/marshall/sthakrar/go/src/github.com/SumilThakr/aqcomp/output/"

	mss := initMs(csvFolder, ncfFolder)
	for _, i := range mss {
		fmt.Println("Getting results for: %s", i.csvPath)
		var tWrt []XY
		results, err := initResults(i)
		if err != nil {
			fmt.Println(err)
		}
		i.results = results
		for _, vals := range i.results {
			tWrt = append(tWrt, XY{vals.simulatedPM, vals.measuredPM})
		}
		errWrite := csvWriter(outputFolder+i.date.Format("20060102")+".csv", tWrt)
		if errWrite != nil {
			log.Fatal(errWrite)
		}

	}

}
