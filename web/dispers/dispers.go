package dispers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
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
	var in dispers.InputCI
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	for _, element := range in.Concepts {
		err := enclave.CreateConcept(&element)
		if err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, dispers.OutputCI{
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

	out := make([]dispers.Concept, len(strConcepts))
	for i, strConcept := range strConcepts {
		tmpConcept := dispers.Concept{IsEncrypted: isEncrypted, Concept: strConcept}
		err := enclave.GetConcept(&tmpConcept)
		if err != nil {
			return err
		}
		out[i] = tmpConcept
	}

	return c.JSON(http.StatusOK, dispers.OutputCI{
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
		tmpConcept := dispers.Concept{IsEncrypted: isEncrypted, Concept: strConcept}
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

	var inputTF dispers.InputTF
	if err := json.NewDecoder(c.Request().Body).Decode(&inputTF); err != nil {
		return err
	}

	finallist, err := enclave.SelectAddresses(inputTF)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dispers.OutputTF{
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
func query(c echo.Context) error {

	var inputT dispers.InputT

	if err := json.NewDecoder(c.Request().Body).Decode(&inputT); err != nil {
		return err
	}

	data, err := enclave.QueryTarget(inputT)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dispers.OutputT{Data: data})
}

/*
*
*
Data Aggegator's ROUTES : those functions are used on route ./dispers/dataaggregator
*
*
*/
func aggregate(c echo.Context) error {
	var in dispers.InputDA

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

// Routes sets the routing for the dispers service
// ":concepts" has to be a list of concepts separated by ":"
func Routes(router *echo.Group) {

	// TODO : Create a route to retrieve public key
	router.GET("/conceptindexor/concept/:concepts/:is-encrypted", getHash)
	router.POST("/conceptindexor/concept", createConcept)
	router.DELETE("/conceptindexor/concept/:concepts/:is-encrypted", deleteConcepts)

	router.POST("/targetfinder/addresses", selectAddresses)

	router.POST("/target/query", query)

	router.POST("/dataaggregator/aggregation", aggregate)
}
