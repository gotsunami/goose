package goose

import (
	"bf/core/config"
	"fmt"
	"strings"
)

type Search struct {
	Id      ScrollId `bson:"_id"`
	Path    string
	Builder *QueryBuilder
}

// A query set defines search criteria used when looking for products or
// boutiques matches in the search engine.
type QuerySet struct {
	orderBy            SortCriteria
	name               string  // boutique name
	btype              int     // boutique type
	title              string  // product title
	category           int     // product category
	price              float64 // product price
	minprice, maxprice float64 // Used instead of price value if any is set
	from               *Location
	radius             uint // Distance in Km around the location
	coordinates        [][2]float64
	limit              uint // Max results to be returned
	offset             uint // Starting offset inside the result search
}

type SortCriteria int

const (
	ByDate      SortCriteria = 1 << iota // Descendant
	ByPrice                              // Ascendant
	ReverseSort                          // Descendant
)

// NewQuerySet defines a new query set.
//
// QuerySet methods are chainable since they always return a pointer to
// QuerySet. Whatever fields are set, only relevant information is used to
// build the final query sent to the search engine, i.e. defining a QuerySet
// with a price then searching for a boutique will have no negative effect on
// result search since the  price value will be ignored in this context.
func NewQuerySet() *QuerySet {
	qs := new(QuerySet)
	qs.category = -1
	qs.price = -1
	qs.minprice = -1
	qs.maxprice = -1
	qs.limit = 10 // ElasticSearch's default value
	qs.coordinates = nil
	return qs
}

func (qs *QuerySet) String() string {
	return fmt.Sprintf("Title: %s, Cat: %s\nPrice: %.2f, Min: %.2f, Max: %.2f\nRadius: %d\nLimit: %d",
		qs.title, qs.category, qs.price, qs.minprice, qs.maxprice, qs.radius, qs.limit)
}

// Title sets a title. Will be used as Title value for a product search
// It will be converted to lowercase at runtime.
func (qs *QuerySet) Title(t string) *QuerySet {
	qs.title = strings.ToLower(t)
	return qs
}

// Name sets a name. Will be used as Name value for a boutique search.
// It will be converted to lowercase at runtime.
func (qs *QuerySet) Name(t string) *QuerySet {
	qs.name = strings.ToLower(t)
	return qs
}

// Type sets which boutique's type to look for.
func (qs *QuerySet) Type(t int) *QuerySet {
	qs.btype = t
	return qs
}

// Category sets which product's category to look for.
func (qs *QuerySet) Category(c int) *QuerySet {
	qs.category = c
	return qs
}

// Price sets the product's price to look for.
func (qs *QuerySet) Price(p float64) *QuerySet {
	qs.price = p
	qs.minprice = p
	qs.maxprice = p
	return qs
}

// MinPrice sets the product's minimum price to look for. When set,
// the price value defined with Price() is ignored.
func (qs *QuerySet) MinPrice(p float64) *QuerySet {
	qs.minprice = p
	return qs
}

// MaxPrice sets the product's maximum price to look for. When set,
// the price value defined with Price() is ignored. If both MinPrice
// and MaxPrice are set, an InvalidQueryError is returned by the Find()
// functions if MaxPrice < MinPrice.
func (qs *QuerySet) MaxPrice(p float64) *QuerySet {
	qs.maxprice = p
	return qs
}

// Area sets a geographical constraint on the result set. Radius is
// expressed in kilometers (Km).
// Set coordinates to nil to avoid ES failure:
// only one of geo_distance and geo_polygon can be executed at
// once in a filtered query
func (qs *QuerySet) Area(from *Location, radius uint) *QuerySet {
	qs.coordinates = nil
	qs.from = from
	qs.radius = radius
	return qs
}

// Area sets a geographical constraint on the result set.
// Region is the index of an array of points (polygon).
// Set from to nil to avoid ES failure:
// only one of geo_distance and geo_polygon can be executed at
// once in a filtered query
func (qs *QuerySet) Polygon(region int) *QuerySet {
	qs.from = nil
	qs.coordinates = config.CountryRegionPolygons[region]
	return qs
}

// SortBy defines a sort mask, which is applied when relevant, i.e. a
// sort by price would only apply if price info is supplied.
func (qs *QuerySet) SortBy(sort SortCriteria) *QuerySet {
	qs.orderBy = sort
	return qs
}

// Limit defines the maximum number of results to be returned.
func (qs *QuerySet) Limit(max uint) *QuerySet {
	qs.limit = max
	return qs
}

// Offset defines the starting offset inside the result set.
func (qs *QuerySet) Offset(max uint) *QuerySet {
	qs.offset = max
	return qs
}
