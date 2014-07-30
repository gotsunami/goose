package goose

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// defines units known by ES
type Unit string
const (
	Meters		= Unit("m")
	Kilometers  = Unit("km")
)

type M map[string]interface{}

// only implements the terms facet, but other facets can be added
type Facet struct {
	Terms M `json:"terms"`
}

type Doc struct {
	Doc interface{} `json:"doc"`
}

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"lon"`
}

type BoundingBox struct {
	TopLeft     Location `json:"top_left"`
	BottomRight Location `json:"bottom_right"`
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
				GeoBoundingBox M `json:"geo_bounding_box"`
				GeoDistance M `json:"geo_distance"`
				GeoPolygon M `json:"geo_polygon"`
			} `json:"filter"`
		} `json:"filtered"`
	} `json:"query"`
	Sort     []M              `json:"sort,omitempty"`
	Facets   map[string]Facet `json:"facets"`
	warnings []error
}

func NewQueryBuilder() *QueryBuilder {
	qb := new(QueryBuilder)
	qb.Size = 10 // ElasticSearch's default value
	qb.warnings = make([]error, 0)
	// Default sorting
	//    qb.Sort = append(qb.Sort, M{"creation": map[string]string{"order":"desc"}})
	return qb
}

func (qb *QueryBuilder) Warnings() []error {
	return qb.warnings
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

// AddGeoDistance defines a distance around a point to filter results
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddGeoDistance("hq.location", Location{Lat:43.454834, Long:3.757789}, 12, Kilometers)
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "match_all": {}
//      },
//      "filter": {
//				"geo_distance": {
//						"distance":"12km",
//						"hq.location": {
//								"lat":43.454834,
//								"lon":3.757789
//						},
//				}
//		},
//		"from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddGeoDistance(name string, point Location, distance uint, unit Unit) *QueryBuilder {
	qb.Query.Filtered.Filter.GeoDistance = M{"distance": fmt.Sprintf("%d%s", distance, unit), name: point}

	return qb
}

// AddGeoBoundingBox defines a bounding box with two points to filter results
// Points are top left and bottom right corners
//
// For example, the following snippet
//  qb := NewQueryBuilder().AddGeoBoundingBox("hq.location", Location{Lat:44, Long:4}, Location{Lat:43, Long:3})
//  r, err := qb.ToJSON()
// will expand to
//  {
//      "query": {
//          "match_all": {}
//      },
//      "filter": {
//				"geo_bounding_box": {
//						"hq.location": {
//								"top_left": {
//										"lat":44,
//										"lon":4
//								},
//								"bottom_right": {
//										"lat":43,
//										"lon":3
//								}
//						}
//				}
//		},
//		"from": 0,
//      "size": 10
//  }
func (qb *QueryBuilder) AddGeoBoundingBox(name string, topleft, bottomright Location) *QueryBuilder {
	if topleft.Lat < bottomright.Lat {
		qb.warnings = append(qb.warnings, errors.New(fmt.Sprintf("invalid bounding box, topleft latitude (%f) is lower than bottomright latitude (%f)", topleft.Lat, bottomright.Lat)))
	}
	if topleft.Long < bottomright.Long {
		qb.warnings = append(qb.warnings, errors.New(fmt.Sprintf("invalid bounding box, topleft longitude (%f) is lower than bottomright longitude (%f)", topleft.Long, bottomright.Long)))
	} 
	qb.Query.Filtered.Filter.GeoBoundingBox = M{name: BoundingBox { topleft, bottomright }}
	
	return qb
}

// ToJSON marshalizes the QueryBuilder structure and returns a suitable JSON query
// string if and only if no warnings were generated during build.
// Usually, queries won't fail if warnings are produced but the reply will probably 
// not be the expected result (see AddGeoBoundingBox)
func (qb *QueryBuilder) ToJSON() (string, error) {
	if len(qb.warnings) > 0 {
		return "", errors.New("ToJSON() refuses to marshal queries with warnings!")
	}
	return qb.ForceToJSON()
}

// ForceToJSON marshalizes the QueryBuilder structure and returns a suitable JSON 
// query string even if warnings were produced during build.
func (qb *QueryBuilder) ForceToJSON() (string, error) {
	query, err := json.Marshal(qb)
	if err != nil {
		return "", err
	}
	q := string(query)
	// Apply a series of patches needed to have a consistent query suitable
	// for ES.
	//
	// Patch 1: patch the query by removing the empty filters
	// ,"filter":{"geo_bounding_box":null,"geo_distance":null,"geo_polygon":null:null}}}
	// which are not valid and causes a NullPointerException in ES.
	q = strings.Replace(q, `"geo_polygon":null`, "", 1)
	q = strings.Replace(q, `"geo_bounding_box":null,`, "", 1)
	q = strings.Replace(q, `"geo_distance":null,`, "", 1)
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

	fmt.Println(q)
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
