package goose

import (
	"net/url"
	"time"
	
	"testing"
)

const (
	geomapping = `{"properties":{"hq":{"type":"geo_point"}}}`
)

func TestMapping(t *testing.T) {
	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)

	time.Sleep(1 * time.Second)

	err := es.SetMappingRawJSON(&DummyObject{}, geomapping)
	if err != nil {
		t.Error("Cannot add mapping:", err)
	}
}

func TestAddMapping(t *testing.T) {
	mb := NewMappingBuilder().AddMapping("hq", TYPE_GEOPOINT)

	r, err := mb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := geomapping
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}

	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)

	time.Sleep(1 * time.Second)

	err = es.SetMapping(&DummyObject{}, mb)
	if err != nil {
		t.Error("Cannot add mapping:", err)
	}

	TestCleanIndex(t)
}

