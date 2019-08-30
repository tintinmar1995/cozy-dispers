package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	dispersErr "github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/cozy/cozy-stack/pkg/jsonapi"
)

// Hosts is the list of servers Cozy-DISPERS available
var Hosts = []url.URL{
	url.URL{
		Scheme: "http",
		Host:   "localhost:8008",
	},
}

const (
	RoleCI        = "conceptindexor"
	RoleTF        = "targetfinder"
	RoleT         = "target"
	RoleDA        = "dataaggregator"
	RoleStack     = "_find"
	RoleConductor = "query"
	ModeSubscribe = "subscribe"
	ModeQuery     = "dispers"
	ModeStack     = "data"
)

func chooseHost() url.URL {
	return Hosts[rand.Intn(len(Hosts))]
}

// ExternalActor structure gives a way to consider every Cozy-DISPERS server and
// communicate with them. Each server can play the role of CI / TF / T / Conductor / DA
type ExternalActor struct {
	Method string
	Status string
	URL    url.URL
	Role   string
	Path   []string
	Outstr string
	Out    []byte
	//OutMeta dispers.Metadata
}

// NewExternalActor initiate a ExternalActor object
func NewExternalActor(role string, mode string) ExternalActor {
	return ExternalActor{
		Path: []string{mode},
		Role: role,
	}
}

func (act *ExternalActor) DefineConductor(url url.URL, queryid string) {
	act.URL = url
	act.URL.Path = strings.Join(append(act.Path, act.Role, queryid), "/")
}

func (act *ExternalActor) DefineDispersActor(job string) {
	act.URL = chooseHost()
	act.URL.Path = strings.Join(append(act.Path, act.Role, job), "/")
}

func (act *ExternalActor) DefineStack(url url.URL) {
	act.URL = url
}

// MakeRequest makes an HTTP request to another DISPERS Actor
func (act *ExternalActor) MakeRequest(method string, token string, input interface{}, body []byte) error {

	var err error
	if input != nil {
		body, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}

	client := http.Client{}
	request, err := http.NewRequest(method, act.URL.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	if len(token) > 0 {
		request.Header.Set("Authorization", token)
	}
	request.Header.Set("Content-Type", "application/json")

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
		act.Status = "404"
		return dispersErr.WrapErrors(dispersErr.ErrRouteNotFound, "")
	}

	if strings.Contains(act.Outstr, "errors") {

		var receivedErrors map[string][]map[string]interface{}
		err := json.Unmarshal(act.Out, &receivedErrors)
		if err != nil {
			return err
		}

		mapError := receivedErrors["errors"][0]
		status, _ := strconv.Atoi(mapError["status"].(string))
		return jsonapi.NewError(status, act.Method+">"+act.URL.String()+" error : "+mapError["detail"].(string))
	}

	var receivedError map[string]interface{}
	err := json.Unmarshal(act.Out, &receivedError)
	if err != nil {
		return err
	}
	indexCode := strings.LastIndex(receivedError["error"].(string), "code")
	if indexCode != -1 {
		act.Status = receivedError["error"].(string)[indexCode+5 : indexCode+8]
	}

	errorMsg := "cozy-dispers: " + act.Method + ">" + act.URL.String() + " error :"
	errorMsg = errorMsg + "\n" + receivedError["error"].(string)
	return errors.New(errorMsg)

}
