package dispers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
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
	var in dispers.InputCI
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
	return c.JSON(http.StatusOK, dispers.OutputCI{
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

	strConcepts := strings.Split(c.Param("concepts"), "-")
	isEncrypted := true
	if c.Param("encrypted") == "false" {
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

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {

	// TODO : Create a route to retrieve public key
	router.GET("/conceptindexor/concept/:concepts/:encrypted", getHash)
	router.POST("/conceptindexor/concept", createConcept)
	router.DELETE("/conceptindexor/concept/:concepts/:encrypted", deleteConcepts)

	router.POST("/targetfinder/addresses", selectAddresses)
}
