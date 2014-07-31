package goose

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

// performs a search of ElasticObjects with the QueryBuilder matching query.
// qb can be nil, in that case the search looks for all indexed objects
// This is the recommended method to make a search.
// The QueryBuilder is easy to use and handles a lot of exceptions that could provoke
// an ES failure
func (se *ElasticSearch) Search(object ElasticObject, qb *QueryBuilder) (*resultSet, error) {
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
// what you are doing and/or if the QueryBuilder is missing the filter you want
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
