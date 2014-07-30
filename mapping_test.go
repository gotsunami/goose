package goose

import (
	"net/url"

	"testing"
)

const (
	geomapping = `{"goose__dummyobject":{"properties":{"hq":{"type":"geo_point"}}}}`
)

func TestMapping(t *testing.T) {
	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)

	err := es.SetMappingRawJSON(&DummyObject{}, geomapping)
	if err != nil {
		t.Error("Cannot add mapping:", err)
	}
}
