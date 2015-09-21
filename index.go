package goose

import (
	"net/http"
)

// creates an index. Before using an index, it is mandatory to send a XPUT request
func (se *ElasticSearch) CreateIndex() error {
	return se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
}

// use _stats command to check that the index exists
func (se *ElasticSearch) IndexExists() (bool, error) {
	resp, err := se.sendRequestAndGetResponse(GET, se.serverUrl+se.basePath+actionStats, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if err == nil {
		return true, nil
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	defer resp.Body.Close()
	return true, nil
}

// silently creates an index if it does not exist. Intents creation only if no error
// was returned by ElasticSearch.IndexExists()
func (se *ElasticSearch) CreateIndexIfNeeded() error {
	exists, err := se.IndexExists()
	if exists == false && err == nil {
		err = se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
	}
	return err
}

// opens an index
func (se *ElasticSearch) OpenIndex() error {
	return se.sendRequest(POST, se.serverUrl+se.basePath+actionOpen, nil)
}

// closes an index (necessary before calling actions like _settings or _mappings)
func (se *ElasticSearch) CloseIndex() error {
	return se.sendRequest(POST, se.serverUrl+se.basePath+actionClose, nil)
}

// deletes an index
func (se *ElasticSearch) DeleteIndex() error {
	return se.sendRequest(DELETE, se.serverUrl+se.basePath, nil)
}
