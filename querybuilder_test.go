package goose

import (
	"testing"
)

func TestAddSort(t *testing.T) {
	qb := NewSearchQueryBuilder().AddSort("name", ORDER_ASC, MODE_DEF)
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}}}},"sort":[{"name":{"order":"asc"}}]}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}

func TestSetTerm(t *testing.T) {
	qb := NewSearchQueryBuilder().SetTerm("name", "montre")
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
	qb := NewSearchQueryBuilder().AddQueryString("name", "my home")
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
	qb := NewSearchQueryBuilder().AddRange("category", 5, 5)
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
	qb := NewSearchQueryBuilder().AddGreaterThanRange("price", 15)
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
	qb := NewSearchQueryBuilder().AddFloatRange("price", 10.5, 16.9)
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
	qb := NewSearchQueryBuilder().AddLesserThanRange("price", 12.99)
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
	qb := NewSearchQueryBuilder().AddGeoDistance("location", Location{Lat:0, Long:0}, 12, KM)
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
	qb := NewSearchQueryBuilder().AddGeoBoundingBox("location", Location{Lat:90, Long:-180}, Location{Lat:-90, Long:180})
	r, err := qb.ToJSON()
	if err != nil {
		t.Error(err.Error())
	}
	should := `{"from":0,"size":10,"query":{"filtered":{"query":{"match_all":{}},"filter":{"geo_bounding_box":{"location":{"top_left":{"lat":90,"lon":-180},"bottom_right":{"lat":-90,"lon":180}}}}}}}`
	if r != should {
		t.Errorf("wrong JSON. Expected\n%v\ngot\n%v", should, r)
	}
}
