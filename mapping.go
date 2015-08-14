package goose

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// Defines the mapping types available in ES
type MappingType string

const (
	TYPE_DATE     = MappingType("date")
	TYPE_GEOPOINT = MappingType("geo_point")
	TYPE_STRING   = MappingType("string")

	TYPE_BYTE    = MappingType("byte")
	TYPE_SHORT   = MappingType("short")
	TYPE_INTEGER = MappingType("integer")
	TYPE_LONG    = MappingType("long")

	TYPE_FLOAT  = MappingType("float")
	TYPE_DOUBLE = MappingType("double")

	TYPE_BOOLEAN = MappingType("binary")

	TYPE_NULL = MappingType("null")
)

// MappingBuilder has helper functions to build JSON query strings
// compatible with the elasticsearch engine to add mappings.
type MappingBuilder struct {
	Properties map[string]M `json:"properties"`
}

// Returns a pointer to a properly initialized MappingBuilder
func NewMappingBuilder() *MappingBuilder {
	return &MappingBuilder{Properties: make(map[string]M, 0)}
}

// AddMapping defines a new mapping for a named field
//
// For example, the following snippet
//  mb := NewMappingBuilder().AddMapping("hq", TYPE_GEOPOINT)
//  r, err := mb.ToJSON()
// will expand to
//  {
//      "hq": {
//          "type": "geo_point"
//		}
//  }
func (mb *MappingBuilder) AddMapping(name string, t MappingType) *MappingBuilder {
	mb.Properties[name] = M{"type": t}
	return mb
}

// ToJSON marshalizes the MappingBuilder structure and returns a suitable JSON query
// string
func (mb *MappingBuilder) ToJSON() (string, error) {
	b, err := json.Marshal(mb)
	if err != nil {
		return "", err
	}
	return string(b), err
}

// sets a mapping for the object
// Caller is responsible for closing and opening index if necessary
// For example, the following snippet
//  mb := NewMappingBuilder().AddMapping("hq", TYPE_GEOPOINT)
//  err := SetMapping(&DummyObject{}, mb)
// will send the following mapping request:
//  {
//      "goose__dummyobject": {
//          "hq": {
//              "type": "geo_point"
//		    }
//		}
//  }
func (se *ElasticSearch) SetMapping(object ElasticObject, m *MappingBuilder) error {
	mapping, err := m.ToJSON()
	if err != nil {
		return err
	}

	return se.SetMappingRawJSON(object, mapping)
}

// sets a mapping for the object.
// Caller is responsible for closing and opening index if necessary
func (se *ElasticSearch) SetMappingRawJSON(object ElasticObject, mapping string) error {
	path, err := buildPath(object)
	if err != nil {
		return err
	}

	body := strings.NewReader(mapping)

	return se.sendRequest(PUT, se.serverUrl+se.basePath+path+actionMappings, body)
}

// gets the current mapping of the object
// TODO: return a MappingResult
func (se *ElasticSearch) GetMapping(object ElasticObject) (string, error) {
	path, err := buildPath(object)
	if err != nil {
		return "", err
	}

	resp, err := se.sendRequestAndGetResponse(GET, se.serverUrl+se.basePath+path+actionMappings, nil)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	return string(bytes), err
}

// deletes the current mapping of the object along with its data
func (se *ElasticSearch) DeleteMappingAndData(object ElasticObject) (string, error) {
	path, err := buildPath(object)
	if err != nil {
		return "", err
	}

	resp, err := se.sendRequestAndGetResponse(DELETE, se.serverUrl+se.basePath+path+actionMapping, nil)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	return string(bytes), err
}
