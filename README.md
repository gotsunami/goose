goose, a golang API for Elastic Search
======================================

goose is an attempt to write a library to easily use the power of elastic search within a golang program.
The coverage of elastic functionalities is currently quite low but allows one to use the basics:
- CRUD operations on indexes
- Add mapping
- Search
- Sort

It takes advantages of golang interface concept to offer flexibility. A really cool feature is the `QueryBuilder`!

Requirements
------------

goose requires what it exists for: golang and elastic search.
The current version of goose works with golang 1.3 and elastic search 1.x

Installation
------------

Use `go get`:

    go get github.com/gotsunami/goose

Powerful query builders!
------------------------

ES formalism is powerful but quite uneasy to learn or to write as it is not very human readable.

The idea of query builder is to create the ES requests without having to actually write the json content of the request. In addition, some request need PUSH, other need PUT or DELETE, query builders know which one to use.

There are two kinds of builders in goose right now detailed in the following sections:
- Mapping builders
- Search builders

ElasticSearch
-------------

As suggested in elastic search documentation, one should open only one instance of a client to communicate with the elastic search server. In goose, this client is the type `ElasticSearch`. It will do most of the job.

In goose, an instance corresponds to an index so when you create an `ElasticSearch`, you must specify the index. You can create an instance as a global variable like that:

    u, err := url.Parse("http://localhost:9200/my_index/")
    es, err := goose.NewElasticSearch(u)

Note: each time `es` appears in the following document, it refers to the global instance of `ElasticSearch`

ElasticObject
-------------

goose also uses an interface called `ElasticObject` for most API calls.
This interface must define a function called `Key()` that can compute an unique key for each instance.
Once your data structure implements `ElasticObject`, you are almost done! (well, almost I said...)

For instance:

    type HQ struct {
	Company  string `json:"company"`
	Country  uint64 `json:"country"`
	Location goose.Location `json:"location"`
    }

    func (hq *HQ) Key() string {
	return fmt.Sprintf("%s_%d", hq.Company, hq.Country)
    }

Indexes
-------

First things first, basic index operations. As a `ElasicSearch` instance matches an unique index, it is pretty straightforward:

      err := es.CreateIndex()
      err = es.CreateIndexIfNeeded()
      err = es.OpenIndex()
      err = es.CloseIndex()
      err = es.DeleteIndex()

Refer to elastic search documentation to know more about indexes: http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/indices.html

CRUD
----

The `ElasticSearch` provides a set of functions for CRUD operations:

    hq := &HQ{Company:"go-tsunami", Country:33, Location:goose.Location{48.865618, 2.370985}}
    err := es.Insert(hq)
    found, err := es.Get(hq)
    hq.Company = "Go Tsunami"
    err = es.Update(hq)
    err = es.Delete(hq)

An additional  `DeleteByQuery` is available to delete a set of objects.

TODO: `UpdateByQuery`

Refer to elastic search documentation for more information or to contribute: http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/docs.html#docs

Mapping
-------

Currently handled mapping types are:
- date
- geo_point
- string
- long

Here is an exemple of how to use a `MappingBuilder` to add a `geo_point` mapping to field Location of type HQ:

     mb := NewMappingBuilder().AddMapping("location", TYPE_GEOPOINT)
     err := es.SetMapping(&HQ{}, mb)

References:
http://www.elasticsearch.org/guide/en/elasticsearch/reference/current/indices-put-mapping.html

Search
------

At last, here we are, ready to make search queries!

The query builder currently handles the following search criteria:
- geo_distance
- geo_boundingbox
- geo_polygon
- should
- must
- facetting
- fuzzy search
- range
- greater or lower than

Here is an example:

	qb = qb.SetTerm("Company", "Go Tsunami")
	qb = qb.AddGeoBoundingBox("location",
		goose.Location{-180, 90}
		goose.Location{180, -90}

	total, err := es.Count(&HQ{})
	if err != nil {
		return nil, err
	}

	qb.Size = total
	results, err := es.Search(&HQ{}, qb)
	if err != nil {
		return nil, err
	}

	for _, match := range results.Hits.Data {
		hq, ok := match.Object.(*HQ)
		if !ok {
			return ids, errors.New(fmt.Sprintf("ElasticSearch returned an invalid object (%v)", match.Src))
		}
		fmt.Println("An HQ was found for Go Tsunami at GPS coordinates %v", hq.Location)
	}


More
----

Source code contains a lot of comments, especially in the query builder. You can find a full example of a dumb application in example/hq.go.

You can play with it:

    go build example/hq.go
    ./hq

If you want to run queries on the created index, you can find it at `http://localhost:9200/hq/main__hq`

Contribute
----------

goose is an ongoing work, we'd be pleased to review any pull request!  