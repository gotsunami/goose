package goose

 import (
	 "fmt"

// 	"testing"
 )


const (
	uri = "http://localhost:9200/"
	index = "gooseindex"
	index2 = "gooseindex2"
	invalidindex = "UPPERCASE"
	dlat = 43.454834
	dlong = 3.757789
)

type DummyObject struct {
	Id			int		 `json:"id"`
	Description string   `json:"description"`
	Len			float64  `json:"len"`
	HQ			Location `json:"hq"`			
}

func (d *DummyObject) Key() string {
	return fmt.Sprintf("%d", d.Id)
}

// var sid ScrollId

// func TestPrepareScanSearch(t *testing.T) {
// 	qs := NewQuerySet().Title("lulu")
// 	var total int
// 	var err error
// 	sid, total, err = Engine.PrepareSearch(boutiqueSearchPath, qs, "")
// 	if err != nil {
// 		t.Errorf("preparing scan search: %s", err.Error())
// 	}
// 	if total != 2 {
// 		t.Errorf("should return 2 boutiques, got %v", total)
// 	}
// 	if len(string(sid)) == 0 {
// 		t.Errorf("empty scroll id, won't be able to get more hits")
// 	}
// }

// func TestScrollSearch(t *testing.T) {
// 	rset, err := Engine.ScrollSearch(sid)
// 	if err != nil {
// 		t.Errorf("can't scroll search: %s", err.Error())
// 	}
// 	if rset.Hits.Total != 2 {
// 		t.Errorf("should have 2 boutiques, got %v", rset.Hits.Total)
// 	}
// }
