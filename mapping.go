package goose

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type MappingType string

const (
	TYPE_DATE	  = MappingType("date")
	TYPE_GEOPOINT = MappingType("geo_point")
	TYPE_STRING   = MappingType("string")
)

type MappingBuilder struct {
	Properties map[string]M `json:"properties"`
}

func NewMappingBuilder(object ElasticObject) *MappingBuilder {
	return &MappingBuilder{Properties: make(map[string]M, 0)}
}

func (mb *MappingBuilder) AddMapping(name string, t MappingType) *MappingBuilder {
	mb.Properties[name] = M{"type": t}
	return mb
}

func (mb *MappingBuilder) ToJSON() (string, error) {
	b, err := json.Marshal(mb)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func (se *ElasticSearch) SetMapping(object ElasticObject, m *MappingBuilder) error {
	mapping, err := m.ToJSON()
	if err != nil {
		return err
	}

	return se.SetMappingRawJSON(object, mapping)
}

// sets a mapping for the object.
// Caller is responsible for closing and opening index if necessary
// TODO: create a MappingBuilder
func (se *ElasticSearch) SetMappingRawJSON(object ElasticObject, mapping string) error {
	path, err := buildPath(object)
	if err != nil {
		return err
	}

	body := strings.NewReader(mapping)

	_, err = se.sendRequest(PUT, se.serverUrl+se.basePath+path+actionMapping, body)
	return err
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
