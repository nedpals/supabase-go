package postgrest_go

import (
	"net/url"
	"testing"
)

func TestQueryRequestBuilder_Constructor(t *testing.T) {
	client := NewClient(url.URL{Scheme: "https", Host: "example.com"})

	path := "/example_table"
	httpMethod := "GET"

	builder := QueryRequestBuilder{
		client:     client,
		path:       path,
		httpMethod: httpMethod,
		json:       nil,
	}

	if builder.path != path {
		t.Errorf("expected path == %s, got %s", path, builder.path)
	}
	if builder.httpMethod != httpMethod {
		t.Errorf("expected httpMethod == %s, got %s", httpMethod, builder.httpMethod)
	}
	if builder.json != nil {
		t.Errorf("expected json == %v, got %v", nil, builder.json)
	}
}
