package enclave

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func init() {
	job.AddWorker(&job.WorkerConfig{
		WorkerType:   "aggregation",
		Concurrency:  runtime.NumCPU(),
		MaxExecCount: 2,
		WorkerFunc:   WorkerDataAggregator,
	})
	job.AddWorker(&job.WorkerConfig{
		WorkerType:   "query_target",
		Concurrency:  runtime.NumCPU(),
		MaxExecCount: 2,
		WorkerFunc:   WorkerQueryTarget,
	})
}

func handleError(err error) error {

	if err != nil {
		fmt.Println(err.Error())
		// Send err to conductor
	}
	return handleError(err)
}

// WorkerQueryTarget is a worker that launch Target's treatment.
func WorkerQueryTarget(ctx *job.WorkerContext) error {

	// Read Input
	queryStack := &query.StackQuery{}
	if err := ctx.UnmarshalMessage(queryStack); err != nil {
		return handleError(err)
	}
	if &queryStack.LocalQuery == nil {
		return handleError(errors.New("An index is required"))
	}

	// Initialize communication with the stack
	stack := network.NewExternalActor(network.RoleStack, network.ModeStack)
	stack.DefineStack(url.URL{
		Scheme: "http",
		Host:   queryStack.Domain,
		Path:   "data/" + queryStack.LocalQuery.Doctype + "/_index",
	})

	// Create new index
	input := map[string]interface{}{
		"index": queryStack.LocalQuery.Index,
		"name":  "idxdelamortquitue",
		"ddoc":  "_design/idxdelamortquitue",
		"type":  "json",
	}

	// Make the request to add index
	err := stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, input, nil)
	if err != nil {
		return handleError(err)
	}

	// Change URL and make request to get data
	var outStack map[string]interface{}
	stack.URL.Path = "data/" + queryStack.LocalQuery.Doctype + "/_find"
	err = stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, queryStack.LocalQuery.FindRequest, nil)
	if err != nil {
		// Catch error caused by an outdated token
		if strings.Contains(strings.ToLower(err.Error()), "expired token") || strings.Contains(strings.ToLower(err.Error()), "invalid jwt token") {
			outStack = map[string]interface{}{
				"docs": []map[string]interface{}{},
				"next": false,
			}
		} else {
			return handleError(err)
		}
	} else {
		err = json.Unmarshal(stack.Out, &outStack)
		if err != nil {
			return handleError(errors.New("Failed to unmarshal data of stack : " + err.Error()))
		}
	}

	for outStack["next"].(bool) != false {
		// TODO : Deal with pagination when target's data overflow limit
	}

	// Decrypt and unmarshal data
	// TODO : Decrypt data when encrypted by the stack

	// Save data in Target's database
	doc, err := query.NewAsyncTask(queryStack.QueryID, query.AsyncQueryTarget, queryStack.NumberOfTargets)
	if err != nil {
		return handleError(errors.New("Failed to create AsyncTask doc : " + err.Error()))
	}
	// Create an []map[string]interface{}
	data := []map[string]interface{}{}
	for _, rowData := range outStack["docs"].([]interface{}) {
		data = append(data, rowData.(map[string]interface{}))
	}
	if err := doc.SetData(data...); err != nil {
		return handleError(errors.New("Failed set data on AsyncTask : " + err.Error()))
	}

	// Count the number of targets
	fetchEveryTargets, err := query.FetchAsyncDataT(queryStack.QueryID)
	if err != nil {
		return handleError(err)
	}

	if fetchEveryTargets["NumberOfTargets"].(int) == queryStack.NumberOfTargets {
		// Send result to Conductor
		out := query.InputPatchQuery{
			OutT: query.OutputT{
				Data:    fetchEveryTargets["Data"].([]map[string]interface{}),
				QueryID: queryStack.QueryID,
			},
			Role: network.RoleT,
		}

		conductor := network.NewExternalActor(network.RoleConductor, network.ModeQuery)
		conductor.DefineConductor(queryStack.ConductorURL, queryStack.QueryID)
		if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
			if conductor.Status == "409" {
				if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
					query.DeleteAsyncDataT(queryStack.QueryID)
					return handleError(err)
				}
			} else {
				query.DeleteAsyncDataT(queryStack.QueryID)
				return handleError(err)
			}
		}
		return query.DeleteAsyncDataT(queryStack.QueryID)
	}

	return nil
}

// WorkerDataAggregator is a worker that launch DataAggregator's treatment.
func WorkerDataAggregator(ctx *job.WorkerContext) error {

	// Read Input
	in := &query.InputDA{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return handleError(err)
	}

	in.TaskMetadata.Arrival = time.Now()

	// Launch Treatment
	res, err := enclave.AggregateData(*in)
	if err != nil {
		return handleError(err)
	}

	in.TaskMetadata.Returning = time.Now()

	// Send result to Conductor
	out := query.InputPatchQuery{
		OutDA: query.OutputDA{
			Results:       res,
			QueryID:       in.QueryID,
			AggregationID: in.AggregationID,
			TaskMetadata:  in.TaskMetadata,
		},
		Role: network.RoleDA,
	}

	conductor := network.NewExternalActor(network.RoleConductor, network.ModeQuery)
	conductor.DefineConductor(in.ConductorURL, in.QueryID)
	if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
		if conductor.Status == "409" {
			if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
				return handleError(err)
			}
		} else {
			return handleError(err)
		}
	}

	return nil
}
