package goose

import (
	"net/url"
	"time"

	"testing"
)


func TestSearch(t *testing.T) {
	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)
	defer es.DeleteIndex()

	dummySet := []DummyObject{
		DummyObject{
			Id:1,
			Description: "Dummy object 1",
			Len:30.18,
		},
		DummyObject{
			Id:2,
			Description: "Dummy object 2",
			Len:40.00,
		},
	}

	for _, dummy := range dummySet {
		if err := es.Insert(&dummy); err != nil {
			t.Error("Cannot insert dummy object: %v", err)
		}
	}
	time.Sleep(1 * time.Second)

	rset, err := es.Search(&dummySet[0], nil)
	if err != nil {
		t.Error("Search fails: %v", err)
	}
	if rset == nil {
		t.Error("Search result is nil")
	}
	if rset.Hits.Total != 2 {
		t.Error("Invalid number of hits, expected", 2, " got", rset.Hits.Total)
	}
}
