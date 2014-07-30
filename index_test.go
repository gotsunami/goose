package goose

import (
	"fmt"
	"net/url"

	"testing"
)

const (
	index = "gooseindex"
	invalidindex = "UPPERCASE"
)

func TestCreateAndDeleteIndex(t *testing.T) {
	u, _ := url.Parse(fmt.Sprintf("http://localhost:9200/%s", index))
	es, err := NewElasticSearch(u)
	if err != nil {
		t.Fatal("Cannot dial ES: %s", err.Error())
	}
	if err = es.CreateIndex(); err != nil {
		t.Error("Cannot create index: %s", err.Error())
	}

	err, created := es.GetOrCreateIndex()
	if err != nil {
		t.Error("Cannot fetch previously created index: %s", err.Error())
	}
	if created != false {
		t.Error("Should have fetch index but created it: %s", err.Error())
	}

	if err = es.DeleteIndex(); err != nil {
		t.Error("Cannot delete index: %s", err.Error())
	}

}

func TestOpenAndCloseIndex(t *testing.T) {
	u, _ := url.Parse(fmt.Sprintf("http://localhost:9200/%s", index))
	es, _ := NewElasticSearch(u)


	if err := es.OpenIndex(); err == nil {
		t.Error("Can open an inexistant index")
	}

	if err := es.CloseIndex(); err == nil {
		t.Error("Can close an inexistant index")
	}

	es.CreateIndex()
	defer es.DeleteIndex()

	if err := es.OpenIndex(); err != nil {
		t.Error("Cannot open index: %v", err.Error())
	}

	if err := es.CloseIndex(); err != nil {
		t.Error("Cannot close index: %v", err.Error())
	}	
}
