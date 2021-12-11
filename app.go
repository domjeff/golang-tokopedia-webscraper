package app

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"

	"go-scrapping/tokopedia"

	"github.com/gocarina/gocsv"
)

func Start() {

	items := tokopedia.GetAllData(100)
	file, err := os.Create("result.csv")
	defer file.Close()
	if err != nil {
		log.Fatalf(err.Error())
	}

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ';'
		return gocsv.NewSafeCSVWriter(writer)
	})

	gocsv.MarshalFile(&items, file)

}

func setHeaders(req *http.Request) {
	req.Header.Set(
		"User-Agent",
		// "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Mobile Safari/537.36",
		"Chrome/96.0.4664.45",
		// "Chrome",
	)

	// return req
}
