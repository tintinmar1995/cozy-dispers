package enclave

import (
	"errors"
	"fmt"
	"runtime"

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
}

func handleError(err error) error {
	if err != nil {
		fmt.Println(err.Error())
		// Send err to conductor
	}
	return err
}

// WorkerDataAggregator is a worker that launch DataAggregator's treatment.
func WorkerDataAggregator(ctx *job.WorkerContext) error {

	// Read Input
	in := &query.InputDA{}
	if err := ctx.UnmarshalMessage(in); err != nil {
		return handleError(err)
	}
	if len(in.Data) == 0 {
		return handleError(errors.New("Worker has to receive Data to compute the aggregation"))
	}

	// Launch Treatment
	res, err := enclave.AggregateData(*in)
	if err != nil {
		return handleError(err)
	}

	// Send result to Conductor
	out := query.InputPatchQuery{
		OutDA: query.OutputDA{
			Results:       res,
			QueryID:       in.QueryID,
			AggregationID: in.AggregationID,
		},
		Role: network.RoleDA,
	}

	conductor := network.NewExternalActor(network.RoleConductor, network.ModeQuery)
	conductor.DefineConductor(in.ConductorURL, in.QueryID)
	if err := conductor.MakeRequest("PATCH", "", out, nil); err != nil {
		return handleError(err)
	}

	return nil
}
