package query

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/cozy/echo"
)

/*
*
*
CONCEPT INDEXOR'S ROUTES : those functions are used on route ./dispers/conceptindexor/
*
*
*/

func createConcept(c echo.Context) error {

	// Get concept from body
	var in query.InputCI
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	for i, element := range in.Concepts {
		err := enclave.CreateConcept(&element)
		if err != nil {
			return err
		}
		in.Concepts[i] = element
	}
	return c.JSON(http.StatusOK, query.OutputCI{
		Hashes: in.Concepts,
	})
}

func getHash(c echo.Context) error {

	strConcepts := strings.Split(c.Param("concepts"), "-")
	isEncrypted := true
	if c.Param("encrypted") == "false" {
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

	strConcepts := strings.Split(c.Param("concepts"), "-")
	isEncrypted := true
	if c.Param("encrypted") == "false" {
		isEncrypted = false
	}
	if len(strConcepts) == 0 {
		return errors.New("Failed to read concept")
	}

	for _, strConcept := range strConcepts {
		tmpConcept := query.Concept{IsEncrypted: isEncrypted, Concept: strConcept}
		err := enclave.DeleteConcept(tmpConcept)
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

	/*
		data, err := enclave.QueryTarget(inputT)
		if err != nil {
			return err
		}
	*/

	// TODO: Launch worker

	return c.JSON(http.StatusOK, echo.Map{"ok": true})
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

	// TODO: Launch worker
	/*
		results, length, err := enclave.AggregateData(in)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, query.OutputDA{
			Results: results,
			Length:  length,
		})
	*/

	return c.JSON(http.StatusOK, echo.Map{"ok": true})
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

	fetched := &query.QueryDoc{}
	err := couchdb.GetDoc(prefixer.ConductorPrefixer, "io.cozy.ml", queryid, fetched)
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

	var in *query.OutputQ

	if err := json.NewDecoder(c.Request().Body).Decode(in); err != nil {
		return err
	}

	conductor, err := enclave.NewConductor(in)
	if err != nil {
		return err
	}

	err = conductor.Lead()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func updateQuery(c echo.Context) error {

	queryid := c.Param("queryid")

	conductor, err := enclave.NewConductorFetchingQueryDoc(queryid)
	if err != nil {
		return err
	}

	var in query.InputPatchQuery
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	if in.Role == network.RoleDA {
		conductor.Query.Layers[in.OutDA.AggregationID[0]+1].Data = append(conductor.Query.Layers[in.OutDA.AggregationID[0]+1].Data, in.OutDA.Results...)
	}
	couchdb.UpdateDoc(prefixer.ConductorPrefixer, &conductor.Query)

	err = conductor.Lead()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func deleteQuery(c echo.Context) error {

	// TODO: Stop workers (contact T/DA for DELETE /jobs/triggers/:trigger-id)
	// TODO: Retrieve Query Doc
	// TODO: Mark doc as aborted

	return c.NoContent(http.StatusNoContent)
}

// Prevent Conductor from waiting indefinitelly results of QueryCozy / DA
func handleQueryError(c echo.Context) error {

	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

// Routes sets the routing for the dispers service
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
	router.POST("/query/report-error", handleQueryError)
	router.POST("/query/:queryid", updateQuery)
	router.POST("/query/:queryid", deleteQuery)

}
