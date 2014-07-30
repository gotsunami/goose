package goose

func (se *ElasticSearch) CreateIndex() error {
	_, err := se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
	return err
}

func (se *ElasticSearch) GetOrCreateIndex() (error, bool) {
	created := false
	_, err := se.sendRequest(GET, se.serverUrl+se.basePath+actionStats, nil)
	if err != nil {
		_, err = se.sendRequest(PUT, se.serverUrl+se.basePath, nil)
		created = true
	}
	return err, created
}

func (se *ElasticSearch) OpenIndex() error {
	_, err := se.sendRequest(POST, se.serverUrl+se.basePath+actionOpen, nil)	
	return err
}

func (se *ElasticSearch) CloseIndex() error {
	_, err := se.sendRequest(POST, se.serverUrl+se.basePath+actionClose, nil)	
	return err
}

func (se *ElasticSearch) DeleteIndex() error {
	_, err := se.sendRequest(DELETE, se.serverUrl+se.basePath, nil)	
	return err
}
