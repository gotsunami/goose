package goose

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"encoding/json"
)

type ElasticObject interface{
	Key() string
}

// caller must ensure that id is unique for each inserted object.
func (se *ElasticSearch) Insert(object ElasticObject) error {
	if se == nil {
		return errors.New("Search Engine has not been initialized")
	}
	t := reflect.TypeOf(object)
	if t == nil {
		return errors.New("Inserted objects cannot be nil")
	}
	path := fmt.Sprintf("%s_%s/", strings.Replace(t.PkgPath(), "/", "_", -1), t.Name())

	if len(path) == 0 {
		return errors.New("Inserted objects cannot be unnamed types.")
	}

	jsondata, err := json.Marshal(object)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(jsondata))

	_, err = se.sendRequest(PUT, se.serverUrl+se.basePath+path+object.Key(), body)	
	return err
}
