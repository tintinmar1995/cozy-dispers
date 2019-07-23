package enclave

import (
	"bytes"
	"encoding/json"
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
}

// WorkerDataAggregator is a worker that launch DataAggregator's treatment.
func WorkerDataAggregator(ctx *job.WorkerContext) error {
	in := &query.InputDA{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return err
	}
	res, err := enclave.AggregateData(*in)
	if err != nil {
		return err
	}

	out := query.OutputDA{
		Results:       res,
		QueryID:       in.QueryID,
		AggregationID: in.AggregationID,
	}
	in.ConductorURL.Path = "dispers/query/" + in.QueryID
	inputPatchQuery := query.InputPatchQuery{
		IsEncrypted: in.IsEncrypted,
		Role:        "dataaggregator",
		OutDA:       out,
	}
	marshaledInputPatchQuery, err := json.Marshal(inputPatchQuery)
	if err != nil {
		return err
	}

	client := http.Client{}
	request, err := http.NewRequest("POST", in.ConductorURL.String(), bytes.NewReader(marshaledInputPatchQuery))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(resp.Body)

	// TODO: Check conductor's answer

	return err
}
