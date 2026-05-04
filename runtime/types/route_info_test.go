package types

import "testing"

func TestRouteInfo_FindDecorator(t *testing.T) {
	r := &RouteInfo{
		Decorators: []Decorator{
			{Name: "auth", Args: []DecoratorArg{{Kind: "string", Value: "bearer"}}},
			{Name: "cache", Kwargs: map[string]DecoratorArg{
				"maxAge": {Kind: "number", Value: "60"},
				"public": {Kind: "bool", Value: "true"},
			}},
		},
	}

	if d := r.Find("auth"); d == nil || d.Args[0].Value != "bearer" {
		t.Fatalf("expected auth decorator, got %+v", d)
	}

	cache := r.Find("cache")
	if cache == nil {
		t.Fatal("expected cache decorator")
	}
	if cache.Kwargs["maxAge"].Value != "60" {
		t.Errorf("maxAge = %q, want 60", cache.Kwargs["maxAge"].Value)
	}
	if cache.Kwargs["public"].Value != "true" {
		t.Errorf("public = %q, want true", cache.Kwargs["public"].Value)
	}

	if r.Find("missing") != nil {
		t.Error("expected nil for missing decorator")
	}
}
