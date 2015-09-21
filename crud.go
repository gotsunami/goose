package goose

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ElasticObject interface {
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
	v := reflect.ValueOf(object)
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
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	jsondata, err := json.Marshal(object)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(jsondata))

	return se.sendRequest(PUT, se.serverUrl+se.basePath+path+object.Key(), body)
}

// BulkInsert indexes several objects at once using the ES bulk API.
func (se *ElasticSearch) BulkInsert(objects []ElasticObject) error {
	if len(objects) == 0 {
		return errors.New("no object to bulk insert")
	}
	path, err := buildPath(objects[0])
	if err != nil {
		return err
	}
	type index struct {
		Id string `json:"_id"`
	}
	// Index action
	type action struct {
		Index index `json:"index"`
	}
	var buf bytes.Buffer
	for _, object := range objects {
		// Index action then source on next line
		for _, d := range []interface{}{&action{index{object.Key()}}, object} {
			jsondata, err := json.Marshal(d)
			if err != nil {
				return err
			}
			_, err = buf.Write(jsondata)
			if err != nil {
				return err
			}
			buf.Write([]byte("\n")) // Required
		}
	}
	return se.sendRequest(POST, se.serverUrl+se.basePath+path+actionBulk, &buf)
}

// updates an element in the index. TODO: check _update
func (se *ElasticSearch) Update(object ElasticObject) error {
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	jsondata, err := json.Marshal(Doc{Doc: object})
	if err != nil {
		return err
	}
	body := strings.NewReader(string(jsondata))

	return se.sendRequest(POST, se.serverUrl+se.basePath+path+strictSlash(object.Key())+actionUpdate, body)
}

// adds an element to the index. Caller must ensure that id is unique for each inserted object.
func (se *ElasticSearch) Get(object ElasticObject) (bool, error) {
	path, err := buildPath(object)
	if err != nil {
		return false, err
	}
	jsondata, err := json.Marshal(object)
	if err != nil {
		return false, err
	}
	body := strings.NewReader(string(jsondata))

	resp, err := se.sendRequestAndGetResponse(GET, se.serverUrl+se.basePath+path+object.Key(), body)
	defer resp.Body.Close()
	if err != nil {
		return false, err
	}

	dec := json.NewDecoder(resp.Body)
	var res = new(result)
	if err = dec.Decode(res); err != nil {
		return false, err
	}

	if res.Found {
		bj, _ := json.Marshal(res.Src)

		err = json.Unmarshal(bj, object)
		return true, err
	}
	return false, nil
}

// deletes an element from the index
func (se *ElasticSearch) Delete(object ElasticObject) error {
	path, err := buildPath(object)
	if err != nil {
		return err
	}
	return se.sendRequest(DELETE, se.serverUrl+se.basePath+path+object.Key(), nil)
}

// deletes objects with a `query`
// http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/docs-delete-by-query.html
type DeletedIndex struct {
	Shards struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
}
type deleteResponse struct {
	Indices interface{} `json:"_indices"`
}

func (se *ElasticSearch) DeleteByQuery(object ElasticObject, q *QueryBuilder) (*DeletedIndex, error) {
	if q == nil {
		return nil, errors.New("Query is not valid")
	}
	path, err := buildPath(object)
	if err != nil {
		return nil, err
	}
	// delete query does not accept from and size, so set them to 0 for `omitempty` to be triggered
	q.Size = 0
	data, err := q.ToJSON()
	if err != nil {
		return nil, err
	}
	body := strings.NewReader(data)
	resp, err := se.sendRequestAndGetResponse(DELETE, se.serverUrl+se.basePath+path+actionQuery, body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("No response from ES server")
	}
	dec := json.NewDecoder(resp.Body)

	dresp := new(deleteResponse)
	err = dec.Decode(dresp)
	v := reflect.ValueOf(dresp.Indices)
	keys := v.MapKeys()

	for _, key := range keys {
		js, _ := json.Marshal(v.MapIndex(key).Interface())
		index := new(DeletedIndex)
		if err = json.Unmarshal(js, index); err == nil {
			return index, nil
		}
	}
	return nil, nil
}
