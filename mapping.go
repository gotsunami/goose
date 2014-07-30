package goose

import (
	"io/ioutil"
	"strings"
)

// sets a mapping for the object. Close index before setting the mapping and reopen it
// afterwards
// TODO: create a MappingBuilder
func (se *ElasticSearch) SetMappingRawJSON(object ElasticObject, mapping string) error {
	path, err := buildPath(object)
	if err != nil {
		return err
	}

	body := strings.NewReader(mapping)

	if err = se.CloseIndex(); err != nil {
		return err
	}
	if _, err = se.sendRequest(PUT, se.serverUrl+se.basePath+path+actionMapping, body); err != nil {
		return err
	}
	return se.OpenIndex()
}

// gets the current mapping of the object
// TODO: return a MappingResult
func (se *ElasticSearch) GetMapping(object ElasticObject) (string, error) {
	path, err := buildPath(object)
	if err != nil {
		return "", err
	}

	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+path+actionMapping, nil)	
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	return string(bytes), err
}
