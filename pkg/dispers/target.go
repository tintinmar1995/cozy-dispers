package enclave

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

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

		url := strings.Join([]string{"http:/", query.Domain, "data/_find"}, "/")
		marshalFindRequest, err := json.Marshal(query.LocalQuery.FindRequest)
		if err != nil {
			return nil, nil
		}
		resp, err := http.Post(url, "application/json", bytes.NewReader(marshalFindRequest))
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

// GetData decrypts instance given by the conductor and build queries
func GetData(in dispers.InputT) ([]map[string]interface{}, error) {

	queries := make([]dispers.Query, len(in.Targets))

	if in.IsEncrypted {
		if err := decryptInputT(&in); err != nil {
			return []map[string]interface{}{}, err
		}
	}

	for index, item := range in.Targets {
		q := buildQuery(item, in.LocalQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
