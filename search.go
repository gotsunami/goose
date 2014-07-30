package goose

import (
	"encoding/json"
	"strings"
)

func (se *ElasticSearch) Search(object ElasticObject, qs string) (*resultSet, error) {
	path, err := buildPath(object)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(qs)
	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+path+actionSearch+se.stype, body)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(resp.Body)

	var rset = new(resultSet)
	err = dec.Decode(rset)
	if err != nil {
		return nil, err
	}
	// TODO: loop over Hits to cast them as ElasticObject
	return rset, nil
}
