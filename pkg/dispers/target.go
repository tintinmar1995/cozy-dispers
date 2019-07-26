package enclave

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func buildQuery(instance query.Instance, localQuery query.LocalQuery) query.Query {
	// TODO : encrypt outputs
	query := query.Query{
		Domain:      instance.Domain,
		LocalQuery:  localQuery,
		TokenBearer: instance.Token.TokenBearer,
	}

	return query
}

func decryptInputT(in *query.InputT) error {
	return nil
}

func retrieveData(in *query.InputT, queries *[]query.Query) ([]map[string]interface{}, error) {

	var data []map[string]interface{}
	var rowsData map[string]interface{}

	for _, query := range *queries {

		stack := network.NewExternalActor(network.RoleStack, network.ModeStack)
		stack.DefineStack(url.URL{
			Scheme: "http",
			Host:   query.Domain,
			Path:   "data/" + query.LocalQuery.Doctype + "_find/",
		})
		err := stack.MakeRequest("POST", query.TokenBearer, query.LocalQuery.FindRequest, nil)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(stack.Out, &rowsData)
		if err != nil {
			return nil, err
		}

		//if(string(body).)
		fmt.Println(rowsData)
		data = append(data, rowsData["docs"].([]map[string]interface{})...)
	}

	return data, nil
}

// QueryTarget decrypts instance given by the conductor and build queries
func QueryTarget(in query.InputT) ([]map[string]interface{}, error) {

	queries := make([]query.Query, len(in.Targets))

	if in.IsEncrypted {
		if err := decryptInputT(&in); err != nil {
			return nil, err
		}
	}

	if len(in.Targets) == 0 {
		return nil, errors.New("Targets is empty")
	}

	var item2instance query.Instance
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
