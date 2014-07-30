package goose

import (
	"net/url"

	"reflect"
	"testing"
	"time"
)

// consts and types are all defined in es_test.go

func TestInsert(t *testing.T) {
	u, _ := url.Parse(uri+index)
	es, _ := NewElasticSearch(u)
	defer es.DeleteIndex()

	dummy := DummyObject{
		Id:1,
		Description: "Dummy object 1",
		Len:30.18,
	}

	if err := es.Insert(&dummy); err != nil {
		t.Error("Cannot insert dummy object: %v", err)
	}

	bogus := DummyObject{
		Id:1,
	}

	found, err := es.Get(&bogus)
	if found == false || err != nil {
		t.Error("Cannot get dummy object:", err)
	}
	if !reflect.DeepEqual(dummy, bogus) {
		t.Error("Found dummy object has incorrect values, expected", dummy, ", got", bogus)
	} 

	dummy.Len += 12.24
	if err := es.Update(&dummy); err != nil {
		t.Error("Cannot update dummy object: %v", err)
	}

	found, err = es.Get(&bogus)
	if found == false || err != nil {
		t.Error("Cannot get dummy object:", err)
	}
	if !reflect.DeepEqual(dummy, bogus) {
		t.Error("Found dummy object has incorrect values, expected", dummy, ", got", bogus)
	} 

	if err := es.Delete(&dummy); err != nil {
		t.Error("Cannot delete dummy object: %v", err)
	}
	time.Sleep(1 * time.Second)
}
