package enclave

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func buildStackQuery(instance query.Instance, localQuery query.LocalQuery) query.StackQuery {
	// TODO : encrypt outputs
	query := query.StackQuery{
		Domain:      instance.Domain,
		LocalQuery:  localQuery,
		TokenBearer: instance.TokenBearer,
	}

	return query
}

func decryptInputT(in *query.InputT) error {
	return nil
}

func retrieveData(in *query.InputT, queriesStack *[]query.StackQuery) ([]map[string]interface{}, error) {

	var data []map[string]interface{}
	var rowsData map[string]interface{}

	for _, queryStack := range *queriesStack {

		stack := network.NewExternalActor(network.RoleStack, network.ModeStack)
		stack.DefineStack(url.URL{
			Scheme: "http",
			Host:   queryStack.Domain,
			Path:   "data/" + queryStack.LocalQuery.Doctype + "/_index",
		})

		if &queryStack.LocalQuery == nil {
			return nil, errors.New("An index is required")
		}

		input := map[string]interface{}{
			"index": queryStack.LocalQuery.Index,
			"name":  "idxdelamortquitue",
			"ddoc":  "_design/idxdelamortquitue",
			"type":  "json",
		}
		err := stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, input, nil)
		if err != nil {
			return nil, err
		}

		stack.URL.Path = "data/" + queryStack.LocalQuery.Doctype + "/_find"
		err = stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, queryStack.LocalQuery.FindRequest, nil)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(stack.Out, &rowsData)
		if err != nil {
			return nil, err
		}

		for _, row := range rowsData["docs"].([]interface{}) {
			data = append(data, row.(map[string]interface{}))
		}

	}

	return data, nil
}

// QueryTarget decrypts instance given by the conductor and build queries
func QueryTarget(in query.InputT) ([]map[string]interface{}, error) {

	queries := make([]query.StackQuery, len(in.Targets))

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
		q := buildStackQuery(item2instance, in.LocalQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
