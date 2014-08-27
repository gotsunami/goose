package goose

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

type Count struct {
	Count		int `json:count`
}

// Performs a "real" count
// returns {"count":10124,"_shards":{"total":5,"successful":5,"failed":0}}
func (se *ElasticSearch) Count(object ElasticObject) (int, error) {
	path, err := buildPath(object)
	if err != nil {
		return 0, err
	}
	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+path+actionCount, nil)

	dec := json.NewDecoder(resp.Body)

	var count = new(Count)
	err = dec.Decode(count)
	if err != nil {
		return 0, err
	}
	return count.Count, nil
}

// Performs a search count
func (se *ElasticSearch) SearchCount(object ElasticObject, qb *SearchQueryBuilder) (*resultSet, error) {
	return se.search(object, qb, typeCount)
}

// Performs a search
func (se *ElasticSearch) Search(object ElasticObject, qb *SearchQueryBuilder) (*resultSet, error) {
	return se.search(object, qb, typeSearch)
}

// performs a search of ElasticObjects with the SearchQueryBuilder matching query and the search
// type given (scan, search, count).
// qb can be nil, in that case the search looks for all indexed objects
// This is the recommended method to make a search.
// The SearchQueryBuilder is easy to use and handles a lot of exceptions that could provoke
// an ES failure
func (se *ElasticSearch) search(object ElasticObject, qb *SearchQueryBuilder, stype string) (*resultSet, error) { 
	se.stype = stype
	var err error
	jsondata := ""
	if qb != nil {
		if jsondata, err = qb.ToJSON(); err != nil {
			return nil, err
		}
	}
	
	return se.SearchRawJSON(object, jsondata)
}

// performs a search with a (supposedly) valid json string.
// It is strongly adviced not to used this method except if you know exactly
// what you are doing and/or if the SearchQueryBuilder is missing the filter you want
func (se *ElasticSearch) SearchRawJSON(object ElasticObject, jsondata string) (*resultSet, error) {
	path, err := buildPath(object)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(jsondata)
	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+path+actionSearch+se.stype, body)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("No response from ES server")
	}
	dec := json.NewDecoder(resp.Body)

	var rset = new(resultSet)
	err = dec.Decode(rset)
	if err != nil {
		return nil, err
	}

	v := reflect.Indirect(reflect.ValueOf(object)).Type()

	for cnt, r := range rset.Hits.Data {
		bj, _ := json.Marshal(r.Src)
		no := reflect.New(v).Interface()
		err = json.Unmarshal(bj, no)
		if err != nil {
			return rset, err
		}
		rset.Hits.Data[cnt].Object = no
	}
	return rset, nil
}
