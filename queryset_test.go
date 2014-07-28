package goose

import (
	"testing"
)

func TestTitle(t *testing.T) {
	qs := NewQuerySet().Title("Lulu")
	if qs.title != "lulu" {
		t.Errorf("Title should be all lowercase, got %v", qs.title)
	}
}

func TestCategory(t *testing.T) {
	qs := NewQuerySet().Category(0)
	if qs.category != 0 {
		t.Errorf("Bad category name, got %v", qs.category)
	}
}

func TestPrice(t *testing.T) {
	qs := NewQuerySet().Price(5)
	if qs.price != 5 || qs.minprice != 5 || qs.maxprice != 5 {
		t.Errorf("Bad price, got %v/%v/%v", qs.minprice, qs.price, qs.maxprice)
	}
}

func TestMinPrice(t *testing.T) {
	qs := NewQuerySet().MinPrice(5)
	if qs.minprice != 5 {
		t.Errorf("Bad minprice, got %v", qs.minprice)
	}
}

func TestMaxPrice(t *testing.T) {
	qs := NewQuerySet().MaxPrice(5)
	if qs.maxprice != 5 {
		t.Errorf("Bad maxprice, got %v", qs.maxprice)
	}
}

func TestArea(t *testing.T) {
	qs := NewQuerySet().Area(&Location{FullAddress: "Paris"}, 10)
	if qs.from.FullAddress != "Paris" || qs.radius != 10 {
		t.Errorf("Bad area, got %v (%v Kms)", qs.from.FullAddress, qs.radius)
	}
}

func TestLimit(t *testing.T) {
	qs := NewQuerySet().Limit(5)
	if qs.limit != 5 {
		t.Errorf("Bad limit, got %v", qs.limit)
	}
}

func TestSortBy(t *testing.T) {
	qs := NewQuerySet().SortBy(ByPrice)
	if qs.orderBy != ByPrice {
		t.Errorf("Bad sort criteria, got %v", qs.orderBy)
	}
}

func TestOffset(t *testing.T) {
	qs := NewQuerySet().Offset(100)
	if qs.offset != 100 {
		t.Errorf("Bad offset, expect 100, got %v", qs.orderBy)
	}
}
