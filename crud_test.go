package goose

import (
	"net/url"

	"reflect"
	"testing"
	"time"
)

// consts and types are all defined in es_test.go
func TestCrudOperations(t *testing.T) {
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
		t.Error("Cannot update dummy object:", err)
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

	found, err = es.Get(&bogus)
	if found == true {
		t.Error("Found deleted object:", bogus)
	}
	time.Sleep(1 * time.Second)


	u, _ = url.Parse(uri+index2)
	es2, _ := NewElasticSearch(u)

	if err := es.Insert(&dummy); err != nil {
		t.Error("Cannot insert dummy object: %v", err)
	}
	if err := es2.Insert(&dummy); err != nil {
		t.Error("Cannot insert dummy object: %v", err)
	}
	dummy.Id = 2
	if err := es.Insert(&dummy); err != nil {
		t.Error("Cannot insert dummy object: %v", err)
	}

	qb := NewQueryBuilder().SetTerm("id", "1")
	_, err = es.DeleteByQuery(&dummy, qb)
	if err != nil {
		t.Error("Cannot delete by query: %v", err)
	}

	found, err = es.Get(&bogus)
	if found == true {
		t.Error("Can find object (should have been deleted by query):", bogus)
	}
	found, err = es2.Get(&bogus)
	if found == false {
		t.Error("Cannot find object (should not have been deleted by query):", bogus)
	}
	bogus.Id = 2
	found, err = es.Get(&bogus)
	if found == false {
		t.Error("Cannot find object (should not have been deleted by query):", bogus)
	}
	time.Sleep(1 * time.Second)


	TestCleanIndex(t)
}
