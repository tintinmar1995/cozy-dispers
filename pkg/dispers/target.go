package enclave

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

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

func decryptInputT(in *query.InputT) ([]query.Instance, query.LocalQuery, error) {

	var targets []query.Instance
	var localQuery query.LocalQuery

	// TODO: Decrypt inputs if encrypted

	if strings.Contains(string(in.EncryptedTargets), "\\\"") {
		in.EncryptedTargets = []byte(strings.ReplaceAll(string(in.EncryptedTargets), "\"", ""))
		in.EncryptedTargets = []byte(strings.ReplaceAll(string(in.EncryptedTargets), "\\", "\""))
	}

	if err := json.Unmarshal(in.EncryptedTargets, &targets); err != nil {
		return targets, localQuery, errors.New("Failed to unmarshal targets : " + err.Error())
	}
	if err := json.Unmarshal(in.EncryptedLocalQuery, &localQuery); err != nil {
		return targets, localQuery, errors.New("Failed to unmarshal local query : " + err.Error())
	}

	return targets, localQuery, nil
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

		// TODO : Decrypt data

		err = json.Unmarshal(stack.Out, &rowsData)
		if err != nil {
			return nil, errors.New("Failed to unmarshal data of stack : " + err.Error())
		}

		for _, row := range rowsData["docs"].([]interface{}) {
			data = append(data, row.(map[string]interface{}))
		}

	}

	return data, nil
}

// QueryTarget decrypts instance given by the conductor and build queries
func QueryTarget(in query.InputT) ([]map[string]interface{}, error) {

	targets, localQuery, err := decryptInputT(&in)
	if err != nil {
		return nil, err
	}

	queries := make([]query.StackQuery, len(targets))

	if len(targets) == 0 {
		return nil, errors.New("Targets is empty")
	}

	for index, target := range targets {
		q := buildStackQuery(target, localQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
