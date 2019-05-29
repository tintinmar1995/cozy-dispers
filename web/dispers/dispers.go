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
		return c.JSON(http.StatusOK, echo.Map{
			"ok":    false,
			"Error": err,
		})
	}

	// Create array of hashes
	hashes := make([]string, len(in.EncryptedConcepts))
	for index, element := range in.EncryptedConcepts {
		out := enclave.CreateConcept(element)
		if out.Error != nil {
			return c.JSON(http.StatusOK, echo.Map{
				"ok":    false,
				"Error": out.Error,
			})
		}
		hashes[index] = out.Hash
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"ok":    true,
		"Error": nil,
		"hash":  hashes,
	})
}

func deleteConcept(c echo.Context) error {
	concept := c.Param("concept")

	err := enclave.DeleteConcept([]byte(concept))

	return c.JSON(http.StatusCreated, echo.Map{
		"ok":    err == nil,
		"Error": err,
	})
}

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {
	// TODO : Create a route to retrieve public key
	router.POST("/conceptindexor/concept", createConcept)            // hash a concept (and save the salt if needed)
	router.DELETE("/conceptindexor/concept/:concept", deleteConcept) // delete a salt in the database

}
