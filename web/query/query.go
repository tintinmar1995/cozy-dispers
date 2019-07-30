package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/dispers"
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

	queryid := c.Param("queryid")

	fetched := &enclave.QueryDoc{}
	err := couchdb.GetDoc(enclave.PrefixerC, "io.cozy.query", queryid, fetched)
	if err != nil {
		return err
	}

	/*
		metas, err := metadata.RetrieveMetadata(queryid)
		if err != nil {
			return nil, *fetched, err
		}
	*/

	return c.JSON(http.StatusOK, echo.Map{
		"Query":             fetched,
		"ExecutionMetadata": nil,
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

func updateQuery(c echo.Context) error {

	queryid := c.Param("queryid")

	var in query.InputPatchQuery
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		fmt.Println(err.Error())
		return err
	}
	if in.Role == network.RoleDA {
		if err := query.SetAsyncTaskAsFinished(queryid, in.OutDA.AggregationID[0], in.OutDA.AggregationID[1]); err != nil {
			fmt.Println(err.Error())
			return err
		}
		if err := query.SetData(queryid, in.OutDA.AggregationID[0], in.OutDA.AggregationID[1], in.OutDA.Results); err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	// Launch a worker to try to resume query
	msg, err := job.NewMessage(query.InputResumeQuery{QueryID: queryid})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = job.System().PushJob(prefixer.DataAggregatorPrefixer, &job.JobRequest{
		WorkerType: "resume-query",
		Message:    msg,
	})
	if err != nil {
		fmt.Println(err.Error())
		return err
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
	router.PATCH("/query/:queryid", updateQuery)
	router.DELETE("/query/:queryid", deleteQuery)

}
