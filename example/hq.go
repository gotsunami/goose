package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gotsunami/goose"
)

var es *goose.ElasticSearch

type HQ struct {
	Company  string         `json:"company"`
	Country  uint64         `json:"country"`
	Location goose.Location `json:"location"`
}

func (hq *HQ) Key() string {
	return fmt.Sprintf("%s_%d", hq.Company, hq.Country)
}

func printErrAndExit(err error, code int) {
	fmt.Println(err)
	os.Exit(code)
}

func main() {
	u, err := url.Parse("http://localhost:9200/hq/")
	if err != nil {
		printErrAndExit(err, 1)
	}

	if es, err = goose.NewElasticSearch(u); err != nil {
		printErrAndExit(err, 2)
	}
	// No need to create index for HQs, NewElasticSearch did it

	mb := goose.NewMappingBuilder().AddMapping("location", goose.TYPE_GEOPOINT).AddMapping("company", goose.TYPE_STRING)
	if err = es.SetMapping(&HQ{}, mb); err != nil {
		printErrAndExit(err, 3)
	}

	hqi := &HQ{"Go Tsunami", 33, goose.Location{48.865618, 2.370985}}
	if err = es.Insert(hqi); err != nil {
		printErrAndExit(err, 4)
	}

	time.Sleep(1)
	hqf := HQ{Company: "Go Tsunami", Country: 33}
	f, err := es.Get(&hqf)
	if err != nil {
		printErrAndExit(err, 5)
	}
	if f == false {
		fmt.Println("inserted object not found")
		os.Exit(5)
	}
	fmt.Println("Go Tsunami HQ inserted at GPS coordinates", hqf.Location)

	qb := goose.NewQueryBuilder().SetTerm("country", "33")

	results, err := es.Search(&HQ{}, qb)
	if err != nil {
		printErrAndExit(err, 7)
	}

	for _, match := range results.Hits.Data {
		hqf, ok := match.Object.(*HQ)
		if !ok {
			fmt.Println("ElasticSearch returned an invalid object", match.Src)
		}
		fmt.Println("An HQ was found for country code 33:", hqf.Company, hqf.Location)
	}
}
