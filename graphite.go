// package graphite implements in a really simple fashion a way to query a graphite instance's HTTP API.
// To learn more about the Graphite HTTP API, see http://graphite.readthedocs.org/en/1.0/url-api.html
package graphite

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Configuration contains the Graphite URL and a http.Client instance that will make calls to Graphite
type Configuration struct {
	URL    *url.URL
	Client *http.Client
}

type ResultSet []Result

type Result struct {
	Target     string
	DataPoints []DataPoint
}

type DataPoint struct {
	Value float64
	Time  time.Time
}

// Graph represents a graph that will be asked to the Graphite API
type Graph struct {
	// Our graphite instance
	Graphite *Configuration

	// Metrics to ask for
	Targets map[string]bool

	// Parameters to give to Graphite
	Parameters map[string]string
}

func NewConfiguration(graphiteUrl string, client *http.Client) (*Configuration, error) {
	u, err := url.ParseRequestURI(graphiteUrl)

	if err != nil {
		return nil, err
	}

	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("graphiteUrl should be an absolute URI containing both scheme and host")
	}

	return &Configuration{URL: u, Client: client}, nil
}

func NewGraph(c *Configuration) *Graph {
	return &Graph{Graphite: c, Targets: make(map[string]bool), Parameters: make(map[string]string)}
}

func (g *Graph) URL() (*url.URL, error) {
	return url.Parse(g.String())
}

// Add a metric that will be rendered by graphite
func (g *Graph) AddTarget(target string) {
	g.Targets[target] = true
}

// Remove a metric from the list
func (g *Graph) RemoveTarget(target string) {
	delete(g.Targets, target)
}

func (g *Graph) AddParameter(key, value string) {
	g.Parameters[key] = value
}

func (g *Graph) RemoveParameter(key string) {
	delete(g.Parameters, key)
}

// Returns graph's URL as a string
func (g *Graph) String() string {

	return g.Graphite.URL.String() + "/render/?" + g.query().Encode()
}

// Transform targets and parameters to an url.Values object ready to be encoded as a query string
func (g *Graph) query() *url.Values {
	query := &url.Values{}
	for k := range g.Targets {
		query.Add("target", k)
	}

	for k, v := range g.Parameters {
		query.Add(k, v)
	}

	return query
}

// Calls graphite API for the graph data and returns a ResultSet object containing the data series
func (g *Graph) Render() (*ResultSet, error) {
	resp, err := g.render(*g.query(), "json")
	if err != nil {
		return nil, err
	}

	type result struct {
		Target     string
		DataPoints [][]interface{}
	}

	var r []result

	buffer, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buffer, &r)
	if err != nil {
		return nil, err
	}
	rs := ResultSet{}
	for _, v := range r {
		result := &Result{Target: v.Target}
		for _, datapoint := range v.DataPoints {
			if nil == datapoint[0] {
				continue
			}
			value, _ := datapoint[0].(float64)
			t, _ := datapoint[1].(float64)

			result.DataPoints = append(result.DataPoints, DataPoint{Value: value, Time: time.Unix(int64(t), 0)})
		}

		rs = append(rs, *result)
	}

	return &rs, nil
}

// Calls Graphite's API
func (g *Graph) render(query url.Values, format string) (*http.Response, error) {
	query.Set("format", format)
	url := g.Graphite.URL.String() + "/render?" + query.Encode()
	resp, err := http.Get(url)
	if err != nil {
		log.Println(url, "Failed to request.", err)

		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Println(url, "Status code", resp.StatusCode)
	}

	return resp, nil
}
