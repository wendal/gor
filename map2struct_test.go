package gor

import (
	"reflect"
	"testing"
)

func Test_Map2Struct(t *testing.T) {
	m := map[string]interface{}{"permalink": "/:title/:year", "latest": 10}
	post := &PostConfig{}
	ToStruct(m, reflect.ValueOf(post))
	if post.Permalink != m["permalink"].(string) {
		t.Fail()
	}
	if post.Latest != m["latest"].(int) {
		t.Fail()
	}
}

func Test_Map2Struct2(t *testing.T) {
	m := map[string]interface{}{"Theme": "facebook", "pages": map[string]interface{}{"permalink": "/wendal"}}
	top := &TopConfig{}
	ToStruct(m, reflect.ValueOf(top))

	if top.Theme != "facebook" {
		t.Fail()
	}
	if top.Pages.Permalink != "/wendal" {
		t.Fail()
	}
}

func Test_Map2Struct3(t *testing.T) {
	m := map[string]interface{}{"title": "wendal", "navigation": []string{"admin.html", "user.html"}, "author": map[string]interface{}{"name": "wendal"}}
	site := &SiteConfig{}
	ToStruct(m, reflect.ValueOf(site))

	if site.Title != "wendal" {
		t.Fail()
	}
	if site.Navigation[0] != "admin.html" {
		t.Fail()
	}
	if site.Navigation[1] != "user.html" {
		t.Fail()
	}
	PrintJson(site)
}
