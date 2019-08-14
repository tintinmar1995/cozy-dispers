package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/metadata"
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

	for _, element := range in.Concepts {
		err := enclave.CreateConcept(&element)
		if err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, query.OutputCI{
		Hashes: in.Concepts,
	})
}

func getHash(c echo.Context) error {

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
		tmpConcept := query.Concept{IsEncrypted: isEncrypted, Concept: strConcept}
		err := enclave.GetConcept(&tmpConcept)
		if err != nil {
			return err
		}
		out[i] = tmpConcept
	}

	return c.JSON(http.StatusOK, query.OutputCI{
		Hashes: out,
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
		tmpConcept := query.Concept{IsEncrypted: isEncrypted, Concept: strConcept}
		err := enclave.DeleteConcept(&tmpConcept)
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
func selectAddresses(c echo.Context) error {

	var inputTF query.InputTF
	if err := json.NewDecoder(c.Request().Body).Decode(&inputTF); err != nil {
		return err
	}

	finallist, err := enclave.SelectAddresses(inputTF)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, query.OutputTF{
		ListOfAddresses: finallist,
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

	data, err := enclave.QueryTarget(inputT)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, query.OutputT{Data: data})
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

func updateQueryDA(c echo.Context) error {

	// Retrieve input
	queryid := c.Param("queryid")
	var in query.InputPatchQuery
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Check if you have not got already those results
	state, err := query.FetchAsyncStateDA(queryid, in.OutDA.AggregationID[0], in.OutDA.AggregationID[1])
	if err != nil {
		return err
	}
	if state == query.Running {
		// This is the first we try to save results
		// We can save it and try to resume query right after
		if err := query.SetAsyncTaskAsFinished(queryid, in.OutDA.AggregationID[0], in.OutDA.AggregationID[1]); err != nil {
			return err
		}
		task := in.OutDA.TaskMetadata
		task.EndTask(nil)
		if err := query.SetData(queryid, in.OutDA.AggregationID[0], in.OutDA.AggregationID[1], in.OutDA.Results, task); err != nil {
			return err
		}
	} else if state == query.Finished {
		// There has been a conflict but there is no problem
		// Results have been successfully saved the first time
		// This thread has to be killed, another one is resuming the query
		return c.JSON(http.StatusOK, echo.Map{"ok": true, "query_id": queryid})
	} else {
		return errors.New("Cannot get results from a DA that has not been launched")
	}

	// AsyncDoc has successfully been updated, now try to resume query
	// As first security to prevent conflict, we check if indexLayer is finished
	// There will be more check-up later in the thread
	queryDoc, err := enclave.NewQueryFetchingQueryDoc(queryid, in.OutDA.AggregationID[0]+1)
	if err != nil {
		return err
	}
	stateLayer, err := query.FetchAsyncStateLayer(queryid, in.OutDA.AggregationID[0], queryDoc.Layers[in.OutDA.AggregationID[0]].Size)
	if err != nil {
		return err
	}
	if stateLayer == query.Finished {
		if err = queryDoc.Lead(); err != nil {
			return err
		}
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

	router.POST("/targetfinder/addresses", selectAddresses)

	router.POST("/target/query", queryCozy)

	router.POST("/dataaggregator/aggregation", aggregate)

	router.GET("/query/:queryid", getQuery)
	router.POST("/query", createQuery)
	router.PATCH("/query/:queryid", updateQueryDA)
	router.DELETE("/query/:queryid", deleteQuery)

}
