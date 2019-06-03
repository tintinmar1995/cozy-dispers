package dispers

import (
	"encoding/json"
	"net/http"

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

	// Create array of hashes
	hashes := make([]string, len(in.EncryptedConcepts))
	var sliceOfMeta []dispers.Metadata
	for index, element := range in.EncryptedConcepts {
		out, metas, err := enclave.CreateConcept(element)
		if err != nil {
			return err
		}

		hashes[index] = out
		sliceOfMeta = append(sliceOfMeta, metas...)
	}
	return c.JSON(http.StatusOK, dispers.OutputCI{
		Hashes: hashes,
		Metadata: sliceOfMeta,
	})
}

func deleteConcept(c echo.Context) error {
	concept := c.Param("concept")

	err := enclave.DeleteConcept([]byte(concept))
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {
	// TODO : Create a route to retrieve public key
	router.POST("/conceptindexor/concept", createConcept)            // hash a concept (and save the salt if needed)
	router.DELETE("/conceptindexor/concept/:concept", deleteConcept) // delete a salt in the database

}
