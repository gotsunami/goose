package goose

import (
	"net/http"
)

// creates an index. Before using an index, it is mandatory to send a XPUT request
func (se *ElasticSearch) CreateIndex() error {
	_, err := se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
	return err
}

// use _stats command to check that the index exists
func (se *ElasticSearch) IndexExists() (bool, error) {
	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+actionStats, nil)
	if err == nil {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

// silently creates an index if it does not exist. Intents creation only if no error
// was returned by ElasticSearch.IndexExists()
func (se *ElasticSearch) CreateIndexIfNeeded() error {
	exists, err := se.IndexExists()
	if exists == false && err == nil {
		_, err = se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
	}
	return err
}

// opens an index
func (se *ElasticSearch) OpenIndex() error {
	_, err := se.sendRequest(POST, se.serverUrl+se.basePath+actionOpen, nil)	
	return err
}

// closes an index (necessary before calling actions like _settings or _mappings)
func (se *ElasticSearch) CloseIndex() error {
	_, err := se.sendRequest(POST, se.serverUrl+se.basePath+actionClose, nil)	
	return err
}

// deletes an index
func (se *ElasticSearch) DeleteIndex() error {
	_, err := se.sendRequest(DELETE, se.serverUrl+se.basePath, nil)	
	return err
}
