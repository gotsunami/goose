package goose

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type M map[string]interface{}

// only implements the terms facet, but other facets can be added
type Facet struct {
	Terms M `json:"terms"`
}

// QueryBuilder has helper functions to build JSON query strings
// compatible with the elasticsearch engine.
type QueryBuilder struct {
	From  int `json:"from"` // Search offset
	Size  int `json:"size"` // Max products in results (limit)
	Query struct {
		Filtered struct {
			Query struct {
				Bool struct {
					Must   []M `json:"must,omitempty"`
					Should []M `json:"should"`
				} `json:"bool"`
			} `json:"query"`
			Filter struct {
				GeoDistance struct {
					distance uint
					Distance string `json:"distance,omitempty"`
					Location struct {
						Lat  float64 `json:"lat,omitempty"`
						Long float64 `json:"lon,omitempty"`
					} `json:"location"`
				} `json:"geo_distance"`
				GeoPolygon struct {
					Location struct {
						Points [][2]float64 `json:"points"`
					} `json:"location"`
				} `json:"geo_polygon"`
			} `json:"filter"`
		} `json:"filtered"`
	} `json:"query"`
	Sort   []M              `json:"sort,omitempty"`
	Facets map[string]Facet `json:"facets"`
}

func NewQueryBuilder() *QueryBuilder {
	qb := new(QueryBuilder)
	qb.Size = 10 // ElasticSearch's default value
	// Default sorting
	//    qb.Sort = append(qb.Sort, M{"creation": map[string]string{"order":"desc"}})
	return qb
}

// SetTermFacet defines a "facet" {"term"} facet
//
// For example, the following snippet
//  qb := NewQueryBuilder().SetTermFacet("facet1", "field1", 50, nil)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//      },
//      "from": 0,
//      "size": 10,
//      "facets": { "facet1": {"terms": {"field": "field1", "size": 50} } }
//  }
func (qb *QueryBuilder) SetTermFacet(facet, field string, size int, terms M) *QueryBuilder {
	if terms == nil {
		terms = make(M, 0)
	}
	terms["field"] = field
	terms["size"] = size
	if qb.Facets == nil {
		qb.Facets = make(map[string]Facet, 0)
	}
	newfacet := Facet{Terms: terms}
	// if qb.Filter.GeoDistance.distance > 0 {
	// 	newfacet.Filter.GeoDistance.Location = qb.Filter.GeoDistance.Location
	// 	newfacet.Filter.GeoDistance.Ranges = append(newfacet.Filter.GeoDistance.Ranges, M{"from":0, "to": qb.Filter.GeoDistance.distance})
	// }
	qb.Facets[facet] = newfacet
	return qb
}

// SetTerm defines a "term" directive inside a "must" condition.
//
// For example, the following snippet
//  qb := NewQueryBuilder().SetTerm("name", "montre")
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {"term": {"name": "montre"}}
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) SetTerm(key, val string) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Must = append(qb.Query.Filtered.Query.Bool.Must, M{"term": map[string]string{key: val}})
	return qb
}

// AddQueryString defines a "query_string" directive inside a "must" condition.
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddQueryString("name", "my home")
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {
//                      "query_string": {
//                          "query": "my home",
//                          "default_field": "name"
//                      }
//                  }
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddQueryString(name, query string) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Must = append(qb.Query.Filtered.Query.Bool.Must, M{"query_string": map[string]string{
		"default_field": name,
		"query":         query,
	}})
	return qb
}

func (qb *QueryBuilder) AddFuzzySearch(name, query string) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Should = append(qb.Query.Filtered.Query.Bool.Should,
		// exact match got a big boost
		M{"match": M{name: map[string]string{
			"boost": "5",
			"query": query,
			"type":  "phrase",
		}}},
		// fuzzy match got a small boost
		M{"match": M{name + ".fuzzy": map[string]string{
			"boost": "1",
			"query": query,
		}}},
	)
	return qb
}

// AddRange defines a "range" constraint on lower and upper limits inside
// a "must" condition, using the "from" and "to" selectors.
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddRange("name", 5, 5)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {
//                      "range": {
//                          "category": {
//                              "to": 5,
//                              "from": 5
//                          }
//                      }
//                  }
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddRange(name string, from, to int) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Must = append(qb.Query.Filtered.Query.Bool.Must, M{"range": map[string]map[string]int{
		name: map[string]int{
			"from": from,
			"to":   to,
		},
	}})
	return qb
}

// AddFloatRange defines a "range" constraint on lower and upper limits inside
// a "must" condition, using the "gte" and "lte" selectors.
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddFloatRange("price", 10.5, 16.9)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {
//                      "range": {
//                          "price": {
//                              "gte": 10.5
//                          }
//                      }
//                  },
//                  {
//                      "range": {
//                          "price": {
//                              "lte": 16.9
//                          }
//                      }
//                  }
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddFloatRange(name string, from, to float64) *QueryBuilder {
	return qb.AddGreaterThanRange(name, from).AddLesserThanRange(name, to)
}

// AddGreaterThanRange defines a lower limit constraint inside
// a "must" condition.
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddGreaterThanRange("price", 15)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {
//                      "range": {
//                          "price": {
//                              "gte": 15
//                          }
//                      }
//                  }
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddGreaterThanRange(name string, from float64) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Must = append(qb.Query.Filtered.Query.Bool.Must, M{"range": map[string]map[string]float64{
		name: map[string]float64{
			"gte": from,
		},
	}})
	return qb
}

// AddLesserThanRange defines a lower limit constraint inside
// a "must" condition.
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddLesserThanRange("price", 12.99)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "bool": {
//              "must": [
//                  {
//                      "range": {
//                          "price": {
//                              "lte": 12.99
//                          }
//                      }
//                  }
//              ]
//          }
//      },
//      "from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddLesserThanRange(name string, to float64) *QueryBuilder {
	qb.Query.Filtered.Query.Bool.Must = append(qb.Query.Filtered.Query.Bool.Must, M{"range": map[string]map[string]float64{
		name: map[string]float64{
			"lte": to,
		},
	}})
	return qb
}

// ToJSON marshalize the QueryBuilder structure and returns a suitable JSON query string.
func (qb *QueryBuilder) ToJSON() (string, error) {
	query, err := json.Marshal(qb)
	if err != nil {
		return "", err
	}
	q := string(query)
	// Apply a series of patches needed to have a consistent query suitable
	// for ES.
	//
	// Patch 1: patch the query by removing the empty filters
	// ,"filter":{"geo_distance":{"location":{}},"geo_polygon":{"location":{"points":null}}}
	// which are not valid and causes a NullPointerException in ES.
	q = strings.Replace(q, `"geo_polygon":{"location":{"points":null}}`, "", 1)
	q = strings.Replace(q, `"geo_polygon":{"location":{"points":[]}}`, "", 1)
	q = strings.Replace(q, `"geo_distance":{"location":{}}`, "", 1)
	q = strings.Replace(q, `{,`, "{", 1)
	q = strings.Replace(q, `,}`, "}", 1)
	q = strings.Replace(q, `,"filter":{}`, "", 1)
	// Patch 2: remove empty should
	q = strings.Replace(q, `,"should":null`, "", 1)
	q = strings.Replace(q, `"should":null`, "", 1)
	q = strings.Replace(q, `,"should":[]`, "", 1)
	q = strings.Replace(q, `"should":[]`, "", 1)
	// Patch 3: replace empty
	// ,"query":{"bool":{}}
	// with a match all query: ,"query":{"match_all":{}}
	q = strings.Replace(q, `"query":{"bool":{}}`, `"query":{"match_all":{}}`, 1)
	// Patch 4: remove useless empty facets
	q = strings.Replace(q, `,"facets":null`, "", 1)
	q = strings.Replace(q, `,"facets":{}`, "", 1)

	return q, nil
}

// Checksum computes a SHA1 sum of the query builder's json
// string representation. Queries with the same search criteria
// have the same checksum.
func (qb *QueryBuilder) Checksum() (string, error) {
	j, err := qb.ToJSON()
	if err != nil {
		return "", err
	}
	s := sha1.New()
	io.WriteString(s, j)
	return fmt.Sprintf("%x", s.Sum(nil)), nil
}

// ParseQuerySet translates the queryset constraints into a suitable data
// structure suitable for later JSON conversion.
func (qb *QueryBuilder) ParseQuerySet(qs *QuerySet) error {
	if qs == nil {
		return errors.New("can't parse nil queryset")
	}
	qb.From = int(qs.offset)
	qb.Size = int(qs.limit)

	if qs.minprice > 0 && qs.maxprice > 0 {
		qb.AddFloatRange("price", qs.minprice, qs.maxprice)
	} else if qs.minprice > 0 {
		qb.AddGreaterThanRange("price", qs.minprice)
	} else if qs.maxprice > 0 {
		qb.AddLesserThanRange("price", qs.maxprice)
	}

	if qs.category > 0 {
		qb.AddRange("category", qs.category, qs.category)
	}

	if qs.btype > 0 {
		qb.AddRange("type", qs.btype, qs.btype)
	}

	if qs.from != nil {
		// Must geocode this location
		// TODO: implement a caching system?
		gc := NewMapQuestGeoCoder()
		gps, err := gc.Geocode(qs.from)
		if err != nil {
			return err
		}
		if gps != nil {
			qb.Query.Filtered.Filter.GeoDistance.distance = qs.radius
			qb.Query.Filtered.Filter.GeoDistance.Distance = fmt.Sprintf("%dkm", qs.radius)
			qb.Query.Filtered.Filter.GeoDistance.Location.Lat = gps.Lat
			qb.Query.Filtered.Filter.GeoDistance.Location.Long = gps.Long
		}
	}

	if qs.coordinates != nil {
		qb.Query.Filtered.Filter.GeoPolygon.Location.Points = qs.coordinates
	}

	switch qs.orderBy {
	case ByDate:
		qb.Sort = append(qb.Sort, M{"creation": map[string]string{"order": "asc"}})
	case ByPrice:
		qb.Sort = append(qb.Sort, M{"price": map[string]string{"order": "asc"}})
	case ByPrice | ReverseSort:
		qb.Sort = append(qb.Sort, M{"price": map[string]string{"order": "desc"}})
	case ByDate | ReverseSort:
		qb.Sort = append(qb.Sort, M{"creation": map[string]string{"order": "desc"}})
	default:
		qb.Sort = append(qb.Sort, M{"creation": map[string]string{"order": "asc"}})
	}
	return nil
}
