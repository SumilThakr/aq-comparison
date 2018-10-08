package main

import (
	"bufio"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

/*
func main() {
	xys, err := readDataConcat("/home/marshall/sthakrar/go/src/github.com/SumilThakr/aqcomp/output")
	if err != nil {
		log.Fatalf("could not read data.txt: %v", err)
	}

	err = plotData("out.pdf", xys)
	if err != nil {
		log.Fatalf("could not plot data: %v", err)
	}

}
*/

func plotData(path string, xys []xy) error {
	// make a file to write to
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", path, err)
	}

	p, err := plot.New()
	if err != nil {
		return fmt.Errorf("could not create plot: %v", err)
	}

	pxys := make(plotter.XYs, len(xys))
	for i, xy := range xys {
		if xy.x > 0 && xy.y > 0 && xy.x < 80 && xy.y < 80 {
			pxys[i].X = xy.x
			pxys[i].Y = xy.y
		}
	}

	s, err := plotter.NewScatter(pxys)
	if err != nil {
		return fmt.Errorf("could not create scatter: %v", err)
	}
	p.Add(s)

	s.GlyphStyle.Shape = draw.CircleGlyph{}
	s.Color = color.RGBA{R: 255, A: 50}
	s.Radius = vg.Points(1)

	l, err := plotter.NewLine(plotter.XYs{
		{0, 0}, {35, 35},
	})
	if err != nil {
		return fmt.Errorf("could not create new line: %v", err)
	}
	p.Add(l)

	// number of pixels and image format
	wt, err := p.WriterTo(4096, 4096, "pdf")
	if err != nil {
		return fmt.Errorf("could not create writer: %v", err)
	}
	_, err2 := wt.WriteTo(f)
	if err2 != nil {
		return fmt.Errorf("could not write to %s: %v:", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close %s: %v", path, err)
	}
	return nil
}

// this is a struct for the data elements.
type xy struct{ x, y float64 }

// we want to make a function that reads the file and puts the data into
// a slice of xys, so they can be used.
func readData(path string) ([]xy, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var xys []xy

	//There is a normal pattern to scanners. Scan the file to s, then do for
	//s.Scan(), and then something to do with s.Text() or s.Bytes(), and
	//check for s.Err().

	s := bufio.NewScanner(f)
	for s.Scan() {
		var x, y float64
		// So for every line, this Sscanf will read the string as this
		// regex of 2 floats separated by a comma. It writes these into
		// x and y.
		_, err := fmt.Sscanf(s.Text(), "%f,%f", &x, &y)
		if err != nil {
			log.Printf("discarding bad data point: %s: %v", s.Text(), err)
		}
		xys = append(xys, xy{x, y})
		// do s.Text() or s.Bytes(), depending on what you want.
		//fmt.Println(s.Text())
	}
	if s.Err() != nil {
		return nil, fmt.Errorf("could not scan: %v", err)
	}
	return xys, nil
}

func writeDataConcat(outputFile string, xys []xy) error {

	var tWrt []XY

	for _, xyFloats := range xys {
		XYStrings := XY{
			strconv.FormatFloat(xyFloats.x, 'f', -1, 64),
			strconv.FormatFloat(xyFloats.y, 'f', -1, 64),
		}
		tWrt = append(tWrt, XYStrings)
	}

	err := csvWriter(outputFile, tWrt)
	if err != nil {
		return fmt.Errorf("Cannot write the concatenated results to file: %v", err)
	}
	return nil
}

func readDataConcat(csvFolder string) ([]xy, error) {
	var csvList []string
	errOne := filepath.Walk(csvFolder, func(path string, info os.FileInfo, err error) error {
		csvList = append(csvList, path)
		return nil
	})
	if errOne != nil {
		log.Fatal(errOne)
	}

	csvList = csvList[1:]

	var xys []xy

	for _, path := range csvList {

		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			var x, y float64
			// So for every line, this Sscanf will read the string as this
			// regex of 2 floats separated by a comma. It writes these into
			// x and y.
			_, err := fmt.Sscanf(s.Text(), "%f,%f", &x, &y)
			if err != nil {
				log.Printf("discarding bad data point: %s: %v", s.Text(), err)
			}
			xys = append(xys, xy{x, y})
			// do s.Text() or s.Bytes(), depending on what you want.
			//fmt.Println(s.Text())
		}
		if s.Err() != nil {
			return nil, fmt.Errorf("could not scan: %v", err)
		}
	}

	errTwo := writeDataConcat("concatResults.csv", xys)
	if errTwo != nil {
		return xys, errTwo
	}

	return xys, nil
}
