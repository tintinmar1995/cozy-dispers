package enclave

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

func buildStackQuery(numberTargets int, conductorURL url.URL, queryid string, instance query.Instance, localQuery query.LocalQuery) query.StackQuery {
	// TODO : encrypt outputs
	query := query.StackQuery{
		Domain:          instance.Domain,
		LocalQuery:      localQuery,
		TokenBearer:     instance.TokenBearer,
		QueryID:         queryid,
		ConductorURL:    conductorURL,
		NumberOfTargets: numberTargets,
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

func retrieveData(in *query.InputT, queriesStack *[]query.StackQuery) error {

	for _, queryStack := range *queriesStack {

		msg, err := job.NewMessage(queryStack)
		if err != nil {
			return err
		}

		_, err = job.System().PushJob(prefixer.TargetPrefixer, &job.JobRequest{
			WorkerType: "query_target",
			Message:    msg,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// QueryTarget decrypts instance given by the conductor and build queries
func QueryTarget(in query.InputT) error {

	targets, localQuery, err := decryptInputT(&in)
	if err != nil {
		return err
	}

	queries := make([]query.StackQuery, len(targets))

	if len(targets) == 0 {
		return errors.New("Targets is empty")
	}

	for index, target := range targets {
		q := buildStackQuery(len(targets), in.ConductorURL, in.QueryID, target, localQuery)
		queries[index] = q
	}

	return retrieveData(&in, &queries)
}
