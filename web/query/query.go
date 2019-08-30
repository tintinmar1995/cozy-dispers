package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/metadata"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/echo"

	// import workers
	_ "github.com/cozy/cozy-stack/worker/dispers"
)

/*
*
*
CONCEPT INDEXOR'S ROUTES : those functions are used on route ./dispers/conceptindexor/
*
*
*/

// createConcept creates concepts in CI's database and returns the corresponding hashes.
func createConcept(c echo.Context) error {

	// Get concept from body
	var in query.InputCI
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	for _, element := range in.EncryptedConcepts {
		err := enclave.CreateConcept(&element, in.IsEncrypted)
		if err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, query.OutputCI{
		Hashes: in.EncryptedConcepts,
	})
}

func getHash(c echo.Context) error {

	meta := metadata.NewTaskMetadata()

	strConcepts := strings.Split(c.Param("concepts"), ":")
	isEncrypted := true
	if c.Param("is-encrypted") == "false" {
		isEncrypted = false
	}
	if len(strConcepts) == 0 {
		return errors.New("Failed to read concept")
	}

	out := make([]query.Concept, len(strConcepts))
	for i, strConcept := range strConcepts {
		tmpConcept := query.Concept{EncryptedConcept: []byte(strConcept)}
		err := enclave.GetConcept(&tmpConcept, isEncrypted)
		if err != nil {
			return err
		}
		out[i] = tmpConcept
	}

	meta.EndTask(nil)

	return c.JSON(http.StatusOK, query.OutputCI{
		Hashes:       out,
		TaskMetadata: meta,
	})
}

func deleteConcepts(c echo.Context) error {

	strConcepts := strings.Split(c.Param("concepts"), ":")
	isEncrypted := true
	if c.Param("is-encrypted") == "false" {
		isEncrypted = false
	}
	if len(strConcepts) == 0 {
		return errors.New("Failed to read concept")
	}

	for _, strConcept := range strConcepts {
		tmpConcept := query.Concept{EncryptedConcept: []byte(strConcept)}
		err := enclave.DeleteConcept(&tmpConcept, isEncrypted)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusNoContent)
}

/*
*
*
TARGET FINDER'S ROUTES : those functions are used on route ./dispers/targetfinder/
*
*
*/
func selectTargets(c echo.Context) error {

	var inputTF query.InputTF
	if err := json.NewDecoder(c.Request().Body).Decode(&inputTF); err != nil {
		return err
	}

	inputTF.TaskMetadata.Arrival = time.Now()

	finallist, err := enclave.SelectAddresses(inputTF)
	if err != nil {
		return err
	}

	// TODO : Encrypt if necessary

	encTargets, err := json.Marshal(finallist)
	if err != nil {
		return err
	}

	inputTF.TaskMetadata.Returning = time.Now()
	return c.JSON(http.StatusOK, query.OutputTF{
		EncryptedTargets: encTargets,
		TaskMetadata:     inputTF.TaskMetadata,
	})
}

/*
*
*
Target'S ROUTES : those functions are used on route ./dispers/target/
*
*
*/
func queryCozy(c echo.Context) error {

	var inputT query.InputT
	if err := json.NewDecoder(c.Request().Body).Decode(&inputT); err != nil {
		return err
	}

	inputT.TaskMetadata.Arrival = time.Now()
	if err := enclave.QueryTarget(inputT); err != nil {
		return err
	}

	inputT.TaskMetadata.Returning = time.Now()
	return c.JSON(http.StatusOK, query.OutputT{
		TaskMetadata: inputT.TaskMetadata,
	})
}

/*
*
*
Data Aggegator's ROUTES : those functions are used on route ./dispers/dataaggregator
*
*
*/
func aggregate(c echo.Context) error {
	var in query.InputDA
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}
	msg, err := job.NewMessage(in)
	if err != nil {
		return err
	}

	_, err = job.System().PushJob(prefixer.DataAggregatorPrefixer, &job.JobRequest{
		WorkerType: "aggregation",
		Message:    msg,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ok": true,
	})
}

/*
*
*
Conducotr'S ROUTES : those functions are used on route ./dispers
*
*
*/
func getQuery(c echo.Context) error {

	queryDoc, err := enclave.NewQueryFetchingQueryDoc(c.Param("queryid"), 0)
	if err != nil {
		return err
	}

	executionMetadata, err := metadata.RetrieveExecutionMetadata(queryDoc.ID())
	if err != nil {
		return err
	}

	asyncMetadata, err := query.FetchAsyncMetadata(queryDoc.ID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"Checkpoints":       queryDoc.CheckPoints,
		"Results":           queryDoc.Results,
		"ExecutionMetadata": executionMetadata,
		"AsyncMetadata":     asyncMetadata,
	})
}

func createQuery(c echo.Context) error {

	var in query.InputNewQuery

	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	query, err := enclave.NewQuery(&in)
	if err != nil {
		return err
	}

	err = query.Lead()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"ok": true, "query_id": query.ID()})
}

func updateQueryDA(in query.OutputDA) error {

	queryid := in.QueryID

	// Check if you have not got already those results
	async, err := query.RetrieveAsyncTaskDA(queryid, in.AggregationID[0], in.AggregationID[1])
	if err != nil {
		return err
	}
	if async.GetStateDA() == query.Running {
		// This is the first we try to save results
		// We can save it and try to resume query right after
		if err := async.SetFinished(); err != nil {
			return err
		}
		task := in.TaskMetadata
		task.EndTask(nil)
		async.TaskMetadata = task
		if err := async.SetData(in.Results); err != nil {
			return err
		}
	} else if async.GetStateDA() == query.Finished {
		// There has been a conflict but there is no problem
		// Results have been successfully saved the first time
		// This thread has to be killed, another one is resuming the query
		return nil
	} else {
		return errors.New("Cannot get results from a DA that has not been launched")
	}

	// AsyncDoc has successfully been updated, now try to resume query
	// As first security to prevent conflict, we check if indexLayer is finished
	// There will be more check-up later in the thread
	queryDoc, err := enclave.NewQueryFetchingQueryDoc(queryid, in.AggregationID[0]+1)
	if err != nil {
		return err
	}
	stateLayer, err := query.FetchAsyncStateLayer(queryid, in.AggregationID[0], queryDoc.Layers[in.AggregationID[0]].Size)
	if err != nil {
		return err
	}
	if stateLayer == query.Finished {
		if err = queryDoc.Lead(); err != nil {
			return err
		}
	}
	return nil
}

func updateQueryT(in query.OutputT) error {

	queryDoc, err := enclave.NewQueryFetchingQueryDoc(in.QueryID, 0)
	if err != nil {
		return err
	}

	queryDoc.Layers[0].Data = in.Data

	if err = queryDoc.Lead(); err != nil {
		return err
	}

	return nil
}

func updateQuery(c echo.Context) error {

	// Retrieve input
	queryid := c.Param("queryid")
	var in query.InputPatchQuery
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		fmt.Println(err.Error())
		return err
	}

	switch in.Role {
	case network.RoleDA:
		if err := updateQueryDA(in.OutDA); err != nil {
			return err
		}
	case network.RoleT:
		if err := updateQueryT(in.OutT); err != nil {
			return err
		}
	default:
		return errors.New("Unknown role")
	}

	return c.JSON(http.StatusOK, echo.Map{"ok": true, "query_id": queryid})
}

func deleteQuery(c echo.Context) error {

	// TODO: Stop workers (contact T/DA for DELETE /jobs/triggers/:trigger-id)
	// TODO: Retrieve Query Doc
	// TODO: Mark doc as aborted

	return c.NoContent(http.StatusNoContent)
}

// Routes sets the routing for the dispers service
// ":concepts" has to be a list of concepts separated by ":"
func Routes(router *echo.Group) {

	// TODO : Create a route to retrieve public key
	router.GET("/conceptindexor/concept/:concepts/:is-encrypted", getHash)
	router.POST("/conceptindexor/concept", createConcept)
	router.DELETE("/conceptindexor/concept/:concepts/:is-encrypted", deleteConcepts)

	router.POST("/targetfinder/addresses", selectTargets)

	router.POST("/target/query", queryCozy)

	router.POST("/dataaggregator/aggregation", aggregate)

	router.GET("/query/:queryid", getQuery)
	router.POST("/query", createQuery)
	router.PATCH("/query/:queryid", updateQuery)
	router.DELETE("/query/:queryid", deleteQuery)

}
