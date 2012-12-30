package graphite

import (
	"net/http"
	"testing"
)

func TestNewConfigurationFailsIfUrlIsNotAbsoluteUri(t *testing.T) {
	array := [...]string{"foobar", "/foobar", "foobar:8080"}

	for _, v := range array {
		c, err := NewConfiguration(v, &http.Client{})
		if c != nil {
			t.Errorf("NewConfiguration should returns a nil instance of Configuration when instanciated with %s", v)
		}
		if err == nil {
			t.Errorf("NewConfiguration should returns a non-nil error when instanciated with %s", v)
		}
	}
}

func TestNewConfigurationReturnsAnInstanceOfConfiguration(t *testing.T) {
	c, err := NewConfiguration("http://foobar.org:8282", &http.Client{})
	if c == nil {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}
}

func TestNewGraph(t *testing.T) {
	c, _ := NewConfiguration("http://foobar:8282", &http.Client{})
	g := NewGraph(c)

	if g == nil {
		t.Fail()
	}
}

func TestGraphStringReturnsGraphAsAString(t *testing.T) {
	c, _ := NewConfiguration("http://foobar:8282", &http.Client{})
	g := NewGraph(c)

	g.AddTarget("foo.bar")

	if g.String() != "http://foobar:8282/render/?target=foo.bar" {
		t.Fail()
	}
}

func TestGraphURLReturnsGraphAsAURL(t *testing.T) {
	c, _ := NewConfiguration("http://foobar:8282", &http.Client{})
	g := NewGraph(c)

	g.AddTarget("foo.bar")

	u, _ := g.URL()
	if u.String() != g.String() {
		t.Fail()
	}
}
