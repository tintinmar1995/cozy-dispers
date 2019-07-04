package enclave

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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

func removeIndex(s *[]dispers.Instance, index int) []dispers.Instance {
	return append((*s)[:index], (*s)[index+1:]...)
}

// cleanTargetsList clean Targets list inplace
func cleanTargetsList(list *[]dispers.Instance) error {

	i := 0
	for i < len(*list) {
		if isInstInvalid(&(*list)[i]) {
			// Unvalid item
			*list = removeIndex(list, i)
			i--
		} else {
			// Valid item
			// Check for duplicates
			j := i + 1
			iHasBeenDeleted := false
			for !iHasBeenDeleted && j < len(*list) {
				if (*list)[i].Domain == (*list)[j].Domain {
					if (*list)[i].SubscriptionDate.After((*list)[j].SubscriptionDate) {
						// Item i is kept
						*list = removeIndex(list, j)
						j--
					} else {
						// Item j is kept
						*list = removeIndex(list, i)
						iHasBeenDeleted = true
					}
				}
				j++
			}
		}
		i++
	}
	// Check missing data
	return nil
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

func isInstInvalid(inst *dispers.Instance) bool {
	if len(inst.Domain) == 0 {
		return true
	}
	if len(inst.Token.TokenBearer) == 0 {
		return true
	}
	if &inst.Domain == nil {
		return true
	}
	if &inst.SubscriptionDate == nil {
		return true
	}
	if inst.SubscriptionDate == (time.Time{}) {
		return true
	}
	if &inst.Token == nil {
		return true
	}
	if &inst.Token.TokenBearer == nil {
		return true
	}
	return false
}

// GetData decrypts instance given by the conductor and build queries
func GetData(in dispers.InputT) ([]map[string]interface{}, error) {

	queries := make([]dispers.Query, len(in.Targets))

	if in.IsEncrypted {
		if err := decryptInputT(&in); err != nil {
			return nil, err
		}
	}

	if err := cleanTargetsList(&in.Targets); err != nil {
		return nil, err
	}

	for index, item := range in.Targets {
		q := buildQuery(item, in.LocalQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
