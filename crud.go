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

// builds the path to an object. If object.BuildPath() string exists, it is called
// else, goose tries to create an index from the object `reflected` data
// Define precisely 
//		object.BuildPath(nil) string
// or buildPath will crash (index out of range or invalid type assertion)
//
// object.BuildPath() must return a string with the following constraints:
// - trailing slash
// - at least two character long (including the trailing slash)
func buildPath(object ElasticObject) (string, error) {
	v := reflect.Indirect(reflect.ValueOf(object))
	t := v.Type()
	if t == nil {
		return "", errors.New("Object cannot be nil")
	}
	
	var path string

	selfBuildPath := v.MethodByName("BuildPath")
	if selfBuildPath.IsValid() {
		valuePath := selfBuildPath.Call(nil)
		// if len(valuePath) != 1 {
		// 	return "", errors.New(fmt.Sprintf("%s.BuildPath() does not return a value.", t.String()))		}
		path = valuePath[0].Interface().(string)
		// if !ok {
		// 	return "", errors.New(fmt.Sprintf("%s.BuildPath() does not return a string (invalid type assertion).", t.String()))		
		// }
		if len(path) < 2 || !strings.HasSuffix(path, "/") {
			return path, errors.New(fmt.Sprintf("%s.BuildPath() returned invalid path.", t.String()))
		}
	} else {
		path = fmt.Sprintf("%s__%s/", strings.Replace(t.PkgPath(), "/", "_", -1), strings.ToLower(t.Name()))
		if len(path) == 2 {
			return path, errors.New("Object cannot be an unnamed type.")
		}
	}
	return path, nil
}

// adds an element to the index. Caller must ensure that id is unique for each inserted object.
func (se *ElasticSearch) Insert(object ElasticObject) error {
	if se == nil {
		return errors.New("Search Engine has not been initialized")
	}
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	jsondata, err := json.Marshal(object)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(jsondata))

	_, err = se.sendRequest(PUT, se.serverUrl+se.basePath+path+object.Key(), body)	
	return err
}

// updates an element in the index.
func (se *ElasticSearch) Update(object ElasticObject) error {
	if se == nil {
		return errors.New("Search Engine has not been initialized")
	}
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	jsondata, err := json.Marshal(object)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(jsondata))

	_, err = se.sendRequest(PUT, se.serverUrl+se.basePath+path+object.Key()+actionUpdate, body)	
	return err
}

// adds an element to the index. Caller must ensure that id is unique for each inserted object.
func (se *ElasticSearch) Get(object ElasticObject) (bool, error) {
	if se == nil {
		return false, errors.New("Search Engine has not been initialized")
	}
	path, err := buildPath(object)
	if err != nil {
		return false, err
	}
	jsondata, err := json.Marshal(object)
	if err != nil {
		return false, err
	}
	body := strings.NewReader(string(jsondata))

	resp, err := se.sendRequest(GET, se.serverUrl+se.basePath+path+object.Key(), body)	
	if err != nil {
		return false, err
	}

	dec := json.NewDecoder(resp.Body)
	var res = new(result)
	if err = dec.Decode(res); err != nil {
		return false, err
	}
	
	if res.Found {
		// v := reflect.Indirect(reflect.ValueOf(object))
		// returnobject := reflect.New(v.Type())

		bj, _ := json.Marshal(res.Src)

		err = json.Unmarshal(bj, object)
		return true, err
	}
	return false, nil
}

// deletes an element from the index
func (se *ElasticSearch) Delete(object ElasticObject) error {
	if se == nil {
		return errors.New("Search Engine has not been initialized")
	}
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	_, err = se.sendRequest(DELETE, se.serverUrl+se.basePath+path+object.Key(), nil)	
	return err
}
