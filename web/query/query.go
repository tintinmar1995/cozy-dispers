package query

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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

	// Get concepts from body
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

	// Convert concepts from string to Concept and collect hashes
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

	// Get concepts from url
	strConcepts := strings.Split(c.Param("concepts"), ":")
	isEncrypted := true
	if c.Param("is-encrypted") == "false" {
		isEncrypted = false
	}
	if len(strConcepts) == 0 {
		return errors.New("Failed to read concept")
	}

	// Convert concepts from string to Concept and collect hashes
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

	// Get Input Data from body (lists of addresses, pseudo-anonymised concepts)
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

	// Read Inputs from body (ListOfAddresses, tokens, ...)
	var inputT query.InputT
	if err := json.NewDecoder(c.Request().Body).Decode(&inputT); err != nil {
		return err
	}

	inputT.TaskMetadata.Arrival = time.Now()

	data, err := enclave.QueryTarget(inputT)
	if err != nil {
		return err
	}

	inputT.TaskMetadata.Returning = time.Now()
	return c.JSON(http.StatusOK, query.OutputT{
		Data:         data,
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

	// Inform that worker has been scheduled
	return c.JSON(http.StatusOK, echo.Map{
		"ok": true,
	})
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

}
