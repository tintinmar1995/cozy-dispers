package enclave

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
)

func buildQuery(instance dispers.Instance, localQuery dispers.LocalQuery) dispers.Query {
	// TODO : encrypt outputs
	query := dispers.Query{
		Domain:      instance.Domain,
		LocalQuery:  localQuery,
		TokenBearer: instance.Token.TokenBearer,
	}

	return query
}

func decryptInputT(in *dispers.InputT) error {
	return nil
}

func retrieveData(in *dispers.InputT, queries *[]dispers.Query) ([]map[string]interface{}, error) {

	var data []map[string]interface{}
	var rowsData []map[string]interface{}

	for _, query := range *queries {

		url := &url.URL{
			Scheme: "http",
			Host:   query.Domain,
			Path:   "data/_find/",
		}

		marshalFindRequest, err := json.Marshal(query.LocalQuery.FindRequest)
		if err != nil {
			return nil, nil
		}
		resp, err := http.Post(url.String(), "application/json", bytes.NewReader(marshalFindRequest))
		if err != nil {
			return nil, nil
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil
		}
		err = json.Unmarshal(body, &rowsData)
		if err != nil {
			return nil, nil
		}

		data = append(data, rowsData...)
	}

	return data, nil
}

// QueryTarget decrypts instance given by the conductor and build queries
func QueryTarget(in dispers.InputT) ([]map[string]interface{}, error) {

	queries := make([]dispers.Query, len(in.Targets))

	if in.IsEncrypted {
		if err := decryptInputT(&in); err != nil {
			return []map[string]interface{}{}, err
		}
	}

	var item2instance dispers.Instance
	for index, item := range in.Targets {
		err := json.Unmarshal([]byte(item), &item2instance)
		if err != nil {
			return nil, err
		}
		q := buildQuery(item2instance, in.LocalQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
