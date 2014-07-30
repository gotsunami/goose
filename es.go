/*
Package cse implements communication with the ElasticSearch engine.

Let's suppose a customer wants to find BMW cars whose price is lesser than 15,000€, only in the
"Cars" category(2), with the results ordered by price from more expensive to less expensive
ones.

    qs := cse.NewQuerySet().MaxPrice(15000).Category(2).SortBy(cse.ByPrice|cse.ReverseSort)
    items, err := Engine.FindItems(qs)
    if err != nil {
        return err
    }

Internally, FindItems() and FindBoutiques() use the QueryBuilder functions to make a suitable,
well-formed JSON object.

    qb := cse.NewQueryBuilder().AddQueryString("name", "bmw")
    qs := cse.NewQuerySet().MaxPrice(15000).Category(2).SortBy(cse.ByPrice|cse.ReverseSort)
    qb.ParseQuerySet(qs)
    r, err := qb.ToJSON()

This will expand to this following JSON object, suitable for searching with ElasticSearch:

    {
        "sort": [
            {
                "price": {
                    "order": "desc"
                }
            }
        ],
        "query": {
            "bool": {
                "must": [
                    {
                        "query_string": {
                            "query": "car",
                            "default_field": "name"
                        }
                    },
                    {
                        "range": {
                            "price": {
                                "lte": 15000
                            }
                        }
                    },
                    {
                        "range": {
                            "category": {
                                "to": 2,
                                "from": 2
                            }
                        }
                    }
                ]
            }
        },
        "from": 0,
        "size": 10
    }
*/
package goose

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	actionSearch       = "_search"
	actionUpdate       = "_update"
	actionOpen		   = "_open"
	actionClose		   = "_close"
	actionStats		   = "_stats"
	typeCount          = "?search_type=count"
	typeScan           = "?search_type=scan&scroll=10m&size=10"
	typeSearch         = "" // Basic search

	envelopeShape  = "envelope"
	withinRelation = "within"
)

const (
	GET    HttpMethod = "GET"
	POST              = "POST"
	PUT               = "PUT"
	DELETE            = "DELETE"
)

var (
	InvalidQueryError  = errors.New("Invalid search engine query.")
	MissingSourceError = errors.New("Missing source in database after a CSE match!")
	jsonStringCleaner  = regexp.MustCompile("(\"|\\|\b|\f|\n|\r|\t|/)")
	strictSlashAdder   = regexp.MustCompile("[/]*$")
)

// removes failing escape chars (see http://json.org/)
// TODO: add \u*
func cleanJsonString(in string) string {
	return jsonStringCleaner.ReplaceAllLiteralString(in, "")
}

func strictSlash(in string) string {
	return strictSlashAdder.ReplaceAllLiteralString(in, "/")
}

type result struct {
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	Id	    string `json:"_id"`
	Version int	   `json:"_version"`
	Found   bool   `json:"found"`
	Src		interface{} `json:"_source"`
}

type resultFacet struct {
	Total int `json:"total"`
	Terms []M `json:"terms"`
}

// Ad-hoc structs for easy unmarshaling search engine's results
type resultSet struct {
	Took int
	Hits struct {
		Total int
		Data  []struct {
			Id  string                 `json:"_id"`
			Src map[string]interface{} `json:"_source"`
		} `json:"hits"`
	}
	Facets map[string]resultFacet `json:"facets"`
}

type scanResultSet struct {
	resultSet
	ScrollId string `json:"_scroll_id"` // query id
}

type ScrollId string

// SearchEngine defines the interface for CRUD operations of our
// central search engine (CSE).
type SearchEngine interface {
	// Item
	// InsertItem(*db.Item, *db.Account) error
	// DeleteItem(*db.Item, *db.Account) error
	// DeleteItemId(string) error
	// BuildItemResultRows(rset *resultSet) ([]*ResultRow, error)
	// CountItems(*QuerySet) (*ItemCount, error)

	// PrepareScanSearch(string, *QuerySet) (ScrollId, int, error)
	// PrepareSearch(string, *QuerySet, ScrollId) (ScrollId, int, error)
	// ScrollSearch(ScrollId) (*resultSet, error)
}

type HttpMethod string

// Global search engine instance.
var Engine *ElasticSearch

// Search engine implementation for elasticsearch.
type ElasticSearch struct {
	serverUrl string
	basePath  string   // defaults to /bf/
	lock      chan bool
	stype     string
}

// NewElasticSearch creates a new ElasticSearch instance which is also
// assigned to the Engine variable. The uri parameter is used to access
// the ElasticSearch web service, i.e http://localhost:9200/<index>
// default search mode is typeSearch
func NewElasticSearch(uri *url.URL) (*ElasticSearch, error) {
	if uri == nil {
		return nil, errors.New("nil ES path")
	}

	// Always set global variable
	Engine := &ElasticSearch{
		serverUrl: uri.Scheme + "://" + uri.Host,
		basePath:  strictSlash(uri.Path),
		lock:      make(chan bool, 1),
		stype:     typeSearch,
	}
	return Engine, Engine.CreateIndexIfNeeded()
}

func (se *ElasticSearch) handleResponse(r *http.Response) error {
	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusCreated {
		d, _ := ioutil.ReadAll(r.Body)
		return errors.New(fmt.Sprintf("HTTP code %d, ES error: %s", r.StatusCode, string(d)))
	}
	return nil
}

// Sends HTTP request to search engine
func (se *ElasticSearch) sendRequest(m HttpMethod, path string, body io.Reader) (*http.Response, error) {
	se.lock <- true
	defer func() { <-se.lock }()
	req, err := http.NewRequest(string(m), path, body)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	err = se.handleResponse(resp)
	return resp, err
}

// PrepareScanSearch initiates a scan search type for scrolling
// efficiently through a large result set. searchPath is the name of the
// category (boutique, item).
//
// The response will include no hits. Instead, the current search's scroll
// id and the number of total hits are returned. The scroll Id can later
// be used with the ScrollSearch function to get some hits.
// func (se *ElasticSearch) PrepareScanSearch(searchPath string, q *QuerySet) (ScrollId, int, error) {
// 	if q == nil {
// 		return "", 0, errors.New("nil query")
// 	}

// 	var err error
// 	var qb *QueryBuilder
// 	if searchPath == "/item/" {
// 		qb, err = NewItemQueryBuilder(q)
// 	} else {
// 		qb, err = NewBoutiqueQueryBuilder(q)
// 	}
// 	if err != nil {
// 		return "", 0, err
// 	}

// 	qs, err := qb.ToJSON()
// 	if err != nil {
// 		return "", 0, err
// 	}

// 	body := strings.NewReader(qs)
// 	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+searchPath+actionSearch+typeScan, body)
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	dec := json.NewDecoder(resp.Body)
// 	var rset = new(scanResultSet)
// 	err = dec.Decode(rset)
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	return ScrollId(rset.ScrollId), rset.Hits.Total, nil
// }

// PrepareSearch initiates a search type for scrolling
// efficiently through a large result set. searchPath is the name of the
// category (boutique, item).
//
// The response will include no hits. Instead, the current search's scroll
// id and the number of total hits is returned. The scroll Id can later
// be used with the ScrollSearch function to get some hits.
// func (se *ElasticSearch) PrepareSearch(searchPath string, q *QuerySet, scid ScrollId) (ScrollId, int, error) {
// 	if q == nil {
// 		return "", 0, errors.New("nil query")
// 	}

// 	var err error
// 	var qb *QueryBuilder
// 	if searchPath == "/item/" {
// 		qb, err = NewItemQueryBuilder(q)
// 	} else {
// 		qb, err = NewBoutiqueQueryBuilder(q)
// 	}
// 	if err != nil {
// 		return "", 0, err
// 	}

// 	qs, err := qb.ToJSON()
// 	if err != nil {
// 		return "", 0, err
// 	}

// 	body := strings.NewReader(qs)
// 	path := se.serverUrl + se.basePath + searchPath + actionSearch
// 	resp, err := se.sendRequest(GET, path+typeScan, body)
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	dec := json.NewDecoder(resp.Body)
// 	var rset = new(scanResultSet)
// 	err = dec.Decode(rset)
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	if len(scid) == 0 {
// 		id := bson.NewObjectId()
// 		scid = ScrollId(id.Hex())
// 	}
// 	_, err = se.c.Upsert(bson.M{"_id": scid}, Search{scid, path, qb})
// 	return scid, rset.Hits.Total, err
// }

// // ScrollSearch retrieves some hits from a previously initiated search
// // request. id is the search scroll id identifying the request returned by
// // the PrepareScanSearch function. The result scroll is complete when no
// // hits have been returned in the resultSet.
// func (se *ElasticSearch) ScrollSearch(id ScrollId) (*resultSet, error) {
// 	search := Search{}
// 	if err := se.c.Find(bson.M{"_id": string(id)}).One(&search); err != nil {
// 		return nil, err
// 	}
// 	qs, err := search.Builder.ToJSON()
// 	if err != nil {
// 		return nil, err
// 	}
// 	body := strings.NewReader(qs)
// 	resp, err := se.sendRequest(GET, search.Path, body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	dec := json.NewDecoder(resp.Body)
// 	var rset = new(resultSet)
// 	err = dec.Decode(rset)
// 	if err != nil {
// 		return nil, err
// 	}
// 	search.Builder.From = search.Builder.From + 10
// 	err = se.c.Update(bson.M{"_id": id}, search)

// 	return rset, err
// }
