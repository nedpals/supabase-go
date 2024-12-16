package postgrest_go

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestBuilder_Constructor(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
	}

	if builder.path != path {
		t.Errorf("expected path == %s, got %s", path, builder.path)
	}
}

func TestRequestBuilder_Select(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
		params: url.Values{},
	}

	s := builder.Select("col1", "col2")

	if got := s.params.Get("select"); got != "col1,col2" {
		t.Errorf("expected param select == %s, got %s", "col1,col2", got)
	}
	if s.httpMethod != "GET" {
		t.Errorf("expected httpMethod == %s, got %s", "GET", s.httpMethod)
	}
	if s.json != nil {
		t.Errorf("expected json == %v, got %v", nil, s.json)
	}
}

func TestRequestBuilder_Insert(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
		header: http.Header{},
		params: url.Values{},
	}

	json := struct{ key1 string }{key1: "val1"}

	s := builder.Insert(json)

	if got := s.header.Get("prefer"); got != "return=representation" {
		t.Errorf("expected param select == %s, got %s", "return=representation", got)
	}
	if s.httpMethod != "POST" {
		t.Errorf("expected httpMethod == %s, got %s", "POST", s.httpMethod)
	}
	if s.json != json {
		t.Errorf("expected json == %v, got %v", json, s.json)
	}
}

func TestRequestBuilder_Upsert(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
		header: http.Header{},
		params: url.Values{},
	}

	json := struct{ key1 string }{key1: "val1"}

	s := builder.Upsert(json)

	if got := s.header.Get("prefer"); got != "return=representation,resolution=merge-duplicates" {
		t.Errorf("expected param select == %s, got %s", "return=representation,resolution=merge-duplicates", got)
	}
	if s.httpMethod != "POST" {
		t.Errorf("expected httpMethod == %s, got %s", "POST", s.httpMethod)
	}
	if s.json != json {
		t.Errorf("expected json == %v, got %v", json, s.json)
	}
}

func TestRequestBuilder_Update(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
		header: http.Header{},
	}

	json := struct{ key1 string }{key1: "val1"}
	s := builder.Update(json)

	if got := s.header.Get("prefer"); got != "return=representation" {
		t.Errorf("expected param select == %s, got %s", "return=representation", got)
	}
	if s.httpMethod != "PATCH" {
		t.Errorf("expected httpMethod == %s, got %s", "PATCH", s.httpMethod)
	}
	if s.json != json {
		t.Errorf("expected json == %v, got %v", json, s.json)
	}
}

func TestRequestBuilder_Delete(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})
	path := "/example_table"

	builder := RequestBuilder{
		client: client,
		path:   path,
	}

	s := builder.Delete()

	if s.httpMethod != "DELETE" {
		t.Errorf("expected httpMethod == %s, got %s", "DELETE", s.httpMethod)
	}
	if s.json != nil {
		t.Errorf("expected json == %v, got %v", nil, s.json)
	}
}
