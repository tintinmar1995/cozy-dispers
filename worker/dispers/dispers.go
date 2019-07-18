package enclave

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
)

// WorkerDataAggregator is a worker that launch DataAggregator's treatment.
func WorkerDataAggregator(ctx *job.WorkerContext) error {
	in := &query.InputDA{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return err
	}
	res, len, err := enclave.AggregateData(*in)
	if err != nil {
		return err
	}

	in.ConductorURL.Path = "dispers/query/" + in.QueryID

	client := http.Client{}
	request, err := http.NewRequest("POST", in.ConductorURL.toString(), bytes.NewReader(dispers.InputPatchQuery{
		IsEncrypted: in.IsEncrypted,
		Role:        "target",
		OutT: dispers.OutputDA{
			Results: res,
			Size:    len,
			QueryId: in.QueryID,
		},
	}))
	if err != nil {
		return err
	}
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType)
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(resp.Body)

	// TODO: Check conductor's answer

	return err
}

func init() {
	job.AddWorker(&job.WorkerConfig{
		WorkerType:   "aggregation",
		Concurrency:  runtime.NumCPU(),
		MaxExecCount: 2,
		WorkerFunc:   WorkerDataAggregator,
	})
}
