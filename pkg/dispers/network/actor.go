package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

// Hosts is the list of servers Cozy-DISPERS available
var Hosts = []url.URL{
	url.URL{
		Scheme: "http",
		Host:   "localhost:8008",
	},
}

const (
	RoleCI = "conceptindexor"
	RoleTF = "targetfinder"
	RoleT  = "target"
	RoleDA = "dataaggregator"
)

func chooseHost() url.URL {
	return Hosts[rand.Intn(len(Hosts))]
}

// ExternalActor structure gives a way to consider every Cozy-DISPERS server and
// communicate with them. Each server can play the role of CI / TF / T / Conductor / DA
type ExternalActor struct {
	Method string
	URL    url.URL
	Role   string
	Mode   string
	Outstr string
	Out    []byte
	//OutMeta dispers.Metadata
}

// SubscribeMode makes an HTTP request to another DISPERS Actor
func (act *ExternalActor) SubscribeMode() {
	act.Mode = "subscribe/"
}

// MakeRequest makes an HTTP request to another DISPERS Actor
func (act *ExternalActor) MakeRequest(method string, job string, contentType string, body []byte) error {

	if len(act.Mode) < 1 {
		act.Mode = "dispers/"
	}

	act.Method = method
	act.URL.Path = strings.Join([]string{act.Mode + act.Role, job}, "/")

	client := http.Client{}
	request, err := http.NewRequest(method, act.URL.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType) // This makes it work
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	act.Outstr = string(body)
	act.Out = body
	if strings.Contains(strings.ToLower(act.Outstr), "error") {
		return act.handleError()
	}

	return nil
}

func (act *ExternalActor) handleError() error {
	if strings.Contains(act.Outstr, "Error") {
		return errors.New("404 : Unknown route")
	}

	var receivedError map[string][]map[string]interface{}
	err := json.Unmarshal(act.Out, &receivedError)
	if err != nil {
		return err
	}

	errorMsg := "cozy-dispers: " + act.Method + ">" + act.URL.String() + " error :"
	for _, mapError := range receivedError["errors"] {
		errorMsg = errorMsg + "\n"
		errorMsg = errorMsg + mapError["status"].(string) + " "
		errorMsg = errorMsg + mapError["detail"].(string)
	}
	return errors.New(errorMsg)
}

// NewExternalActor initiate a ExternalActor object
func NewExternalActor(role string) ExternalActor {
	return ExternalActor{
		URL:  chooseHost(),
		Role: role,
	}
}

// NewSliceOfExternalActors initiate a slice of ExternalActor objects
func NewSliceOfExternalActors(role string, size int) []ExternalActor {
	out := make([]ExternalActor, size)
	for i := range out {
		out[i] = NewExternalActor(role)
	}
	return out
}
