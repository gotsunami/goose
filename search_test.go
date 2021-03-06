package goose

import (
	"net/url"
	"reflect"
	"time"

	"testing"
)

// Exemples of curl commands
//
// curl -XPUT http://localhost:9200/gooseindex;echo
//
// curl -XPOST http://localhost:9200/gooseindex/_close;echo
//
// curl -XPUT  http://localhost:9200/gooseindex/goose__dummyobject/_mapping -d '{"goose__dummyobject":{"properties":{"hq":{"type":"geo_point"}}}}';echo
//
// curl -XPOST http://localhost:9200/gooseindex/_open;echo
//
// curl -XPUT http://localhost:9200/gooseindex/goose__dummyobject/1/ -d '{"id":1,"description":"test","hq":{"lat":10,"lon":10}}';echo
//
// curl -XGET http://localhost:9200/gooseindex/goose__dummyobject/_search;echo
//
// curl -XGET http://localhost:9200/gooseindex/goose__dummyobject/_search -d '{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}},"filter":{"geo_bounding_box":{"hq":{"top_left":{"lat":50,"lon":50},"bottom_right":{"lat":10,"lon":10}}}}}}}';echo

// curl -XGET http://localhost:9200/7bf4178cf48583e6e9a9b4c5411bd71602709ab2/keeneye_api__corner/_search -d '{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}},"filter":{"geo_bounding_box":{"location":{"top_left":{"lat":1000000,"lon":1000000},"bottom_right":{"lat":10,"lon":10}}}}}}}';echo

var dummySet []DummyObject = []DummyObject{
	DummyObject{
		Id:          1,
		Description: "Dummy object 1",
		Len:         30.18,
		HQ:          Location{40.12, -71.34},
	},
	DummyObject{
		Id:          2,
		Description: "My object id is 2",
		Len:         40.00,
		HQ:          Location{dlat, dlong},
	},
}

func TestInitSearch(t *testing.T) {
	u, _ := url.Parse(uri + index)
	es, _ := NewElasticSearch(u)
	time.Sleep(1 * time.Second)
	es.SetMappingRawJSON(&DummyObject{}, geomapping)
}

func TestSearch(t *testing.T) {
	u, _ := url.Parse(uri + index)
	es, _ := NewElasticSearch(u)

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
	} else if rset.Hits.Total != 2 {
		t.Error("Invalid number of hits, expected", 2, " got", rset.Hits.Total)
	}

	var found bool
	for _, dummy := range dummySet {
		found = false
		for _, r := range rset.Hits.Data {
			if reflect.DeepEqual(&dummy, r.Object) {
				found = true
				break
			}
		}
		if found == false {
			t.Error("Dummy object was not found in results: ", dummy)
		}
	}
}

func TestSearchWithAddQueryString(t *testing.T) {
	u, _ := url.Parse(uri + index)
	es, _ := NewElasticSearch(u)

	qb := NewQueryBuilder().AddQueryString("description", "Dummy")

	rset, err := es.Search(&dummySet[0], qb)
	if err != nil {
		t.Error("Search fails: %v", err)
	}
	if rset == nil {
		t.Error("Search result is nil")
	} else if rset.Hits.Total != 1 {
		t.Error("Invalid number of hits, expected", 1, " got", rset.Hits.Total)
	}

	qb = NewQueryBuilder().AddQueryString("description", "object")

	rset, err = es.Search(&dummySet[0], qb)
	if err != nil {
		t.Error("Search fails: %v", err)
	}
	if rset == nil {
		t.Error("Search result is nil")
	} else if rset.Hits.Total != 2 {
		t.Error("Invalid number of hits, expected", 2, " got", rset.Hits.Total)
	}
}

func TestSearchWithAddGeoBoundingBox(t *testing.T) {
	u, _ := url.Parse(uri + index)
	es, _ := NewElasticSearch(u)

	qb := NewQueryBuilder().AddGeoBoundingBox("hq", Location{90, -180}, Location{-90, 1})

	rset, err := es.Search(&dummySet[0], qb)
	if err != nil {
		t.Fatal("Search fails: %v", err)
	}
	if rset == nil {
		t.Error("Search result is nil")
	} else if rset.Hits.Total != 1 {
		t.Error("Invalid number of hits, expected", 1, " got", rset.Hits.Total)
	}
}

func TestCleanIndex(t *testing.T) {
	u, _ := url.Parse(uri + index)
	es, _ := NewElasticSearch(u)
	es.DeleteIndex()
}
