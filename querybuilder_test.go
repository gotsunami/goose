package goose

import (
	"testing"
)

func TestSetTerm(t *testing.T) {
	qb := NewQueryBuilder().SetTerm("name", "montre")
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"term":{"name":"montre"}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddQueryString(t *testing.T) {
	qb := NewQueryBuilder().AddQueryString("name", "my home")
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"query_string":{"default_field":"name","query":"my home"}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddRange(t *testing.T) {
	qb := NewQueryBuilder().AddRange("category", 5, 5)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"range":{"category":{"from":5,"to":5}}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddGreaterThanRange(t *testing.T) {
	qb := NewQueryBuilder().AddGreaterThanRange("price", 15)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"range":{"price":{"gte":15}}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddFloatRange(t *testing.T) {
	qb := NewQueryBuilder().AddFloatRange("price", 10.5, 16.9)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"range":{"price":{"gte":10.5}}},{"range":{"price":{"lte":16.9}}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddLesserThanRange(t *testing.T) {
	qb := NewQueryBuilder().AddLesserThanRange("price", 12.99)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"bool":{"must":[{"range":{"price":{"lte":12.99}}}]}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddGeoDistance(t *testing.T) {
	qb := NewQueryBuilder().AddGeoDistance("location", Location{Lat:0, Long:0}, 12, Kilometers)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}},"filter":{"geo_distance":{"distance":"12km","location":{"lat":0,"lon":0}}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestAddGeoBoundingBox(t *testing.T) {
	qb := NewQueryBuilder().AddGeoBoundingBox("location", Location{Lat:10, Long:10}, Location{Lat:0, Long:0})
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}},"filter":{"geo_bounding_box":{"location":{"top_left":{"lat":10,"lon":10},"bottom_right":{"lat":0,"lon":0}}}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}
