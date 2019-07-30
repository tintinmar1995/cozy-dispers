package enclave

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
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
		WorkerType:   "resume-query",
		Concurrency:  1, // Prevent that conductor computes twice the same task
		MaxExecCount: 2,
		WorkerFunc:   WorkerResumeQuery,
	})
}

func handleError(err error) error {
	if err != nil {
		fmt.Println(err.Error())
		// Send err to conductor
	}
	return err
}

func WorkerResumeQuery(ctx *job.WorkerContext) error {

	in := &query.InputResumeQuery{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return handleError(err)
	}

	queryDoc, err := enclave.NewQueryFetchingQueryDoc(in.QueryID)
	if err != nil {
		return handleError(err)
	}
	if err = queryDoc.Lead(); err != nil {
		return handleError(err)
	}
	if err = queryDoc.TryToEndQuery(); err != nil {
		return handleError(err)
	}

	return nil
}

// WorkerDataAggregator is a worker that launch DataAggregator's treatment.
func WorkerDataAggregator(ctx *job.WorkerContext) error {
	in := &query.InputDA{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return handleError(err)
	}

	if len(in.Data) == 0 {
		return handleError(errors.New("Worker has to receive Data to compute the aggregation"))
	}

	res, err := enclave.AggregateData(*in)
	if err != nil {
		return handleError(err)
	}

	// TODO: Use External Actor

	out := query.OutputDA{
		Results:       res,
		QueryID:       in.QueryID,
		AggregationID: in.AggregationID,
	}
	in.ConductorURL.Path = "dispers/query/" + in.QueryID
	marshaledOutputDA, err := json.Marshal(out)
	if err != nil {
		return handleError(err)
	}
	client := http.Client{}
	request, err := http.NewRequest("PATCH", in.ConductorURL.String(), bytes.NewReader(marshaledOutputDA))
	if err != nil {
		return handleError(err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return handleError(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	// TODO: Check conductor's answer

	return handleError(err)
}
