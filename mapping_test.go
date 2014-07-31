package goose

import (
	"net/url"
	"time"
	
	"testing"
)

const (
	geomapping = `{"properties":{"hq":{"type":"geo_point"}}}`
)

func TestAddMapping(t *testing.T) {
	mb := NewMappingBuilder(&DummyObject{}).AddMapping("hq", TYPE_GEOPOINT)

	r, err := mb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"properties":{"hq":{"type":"geo_point"}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestMapping(t *testing.T) {
	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)

	time.Sleep(1 * time.Second)

	err := es.SetMappingRawJSON(&DummyObject{}, geomapping)
	if err != nil {
		t.Error("Cannot add mapping:", err)
	}
}
