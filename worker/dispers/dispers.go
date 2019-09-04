package enclave

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/metadata"
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
	return err
}

// WorkerQueryTarget is a worker that launch Target's treatment.
func WorkerQueryTarget(ctx *job.WorkerContext) error {

	// Create a task metadata
	task := metadata.NewTaskMetadata()

	// Read Input
	queryStack := &query.StackQuery{}
	if err := ctx.UnmarshalMessage(queryStack); err != nil {
		return handleError(err)
	}
	if &queryStack.LocalQuery == nil {
		return handleError(errors.New("An index is required"))
	}

	// Success or not, we have to sent an array of data to the conductor
	// Query should not stop because of a particular isntance that ruine everything
	// Conductor can decide to stop the query if there is not enough data
	var outStack map[string]interface{}
	var processError error

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

	// Create variable to save received data
	data := []map[string]interface{}{}
	processError = stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, input, nil)

	// Deal with pagination when target's data overflow limit
	pagination := 0
	for pagination == 0 || outStack["next"].(bool) == true {
		if processError == nil {
			// If we succeeded to add a new index
			// Change URL and make request to get data
			queryStack.LocalQuery.FindRequest.Skip = pagination * queryStack.LocalQuery.FindRequest.Limit
			stack.URL.Path = "data/" + queryStack.LocalQuery.Doctype + "/_find"
			processError = stack.MakeRequest("POST", "Bearer "+queryStack.TokenBearer, queryStack.LocalQuery.FindRequest, nil)
		}

		if processError == nil {
			// If we succeded to add an index and to make the request
			// Unmarshal outputs to get data
			processError = json.Unmarshal(stack.Out, &outStack)
			for _, rowData := range outStack["docs"].([]interface{}) {
				data = append(data, rowData.(map[string]interface{}))
			}
		}
		pagination = pagination + 1
	}

	fmt.Println(processError)

	// Decrypt and unmarshal data

	// Save data in Target's database
	// Step 1, create a new AsyncTask doc
	doc, err := query.NewAsyncTask(queryStack.QueryID, query.AsyncQueryTarget, queryStack.NumberOfTargets)
	if err != nil {
		return handleError(errors.New("Failed to create AsyncTask doc : " + err.Error()))
	}
	// Step 2 : end task
	task.EndTask(processError)
	doc.TaskMetadata = task
	// Step 3 : Create an []map[string]interface{} and set data
	// This step update the doc in database and save the TaskMetadata from step 2
	if err := doc.SetData(data...); err != nil {
		return handleError(errors.New("Failed set data on AsyncTask : " + err.Error()))
	}

	// Now we try to figure out if every target has been contact to fetch data
	// Counting the number of targets
	fetchEveryTargets, err := query.FetchAsyncDataT(queryStack.QueryID)
	if err != nil {
		return handleError(errors.New("Failed fetch targets : " + err.Error()))
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

		// Contact the conductor to send data and continue the query
		conductor := network.NewExternalActor(network.RoleConductor, network.ModeQuery)
		conductor.DefineConductor(queryStack.ConductorURL, queryStack.QueryID)
		if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
			if conductor.Status == "409" {
				// If conflict, we try a second time
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
