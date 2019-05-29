package dispers

import (
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

	// TODO: Get concept from body

	// TODO: Decrypt concepts

	// TODO: Create array of hashes

	// TODO: Call HashMeThat for each concept
	hash, err := enclave.HashMeThat(concept)

	return c.JSON(http.StatusCreated, echo.Map{
		"ok":   err == nil,
		"hash": hash,
	})
}

func deleteConcept(c echo.Context) error {
	concept := c.Param("concept")
	err := enclave.DeleteConcept(concept)
	return c.JSON(http.StatusCreated, echo.Map{
		"ok": err == nil,
	})
}

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {
	// TODO : Create a route to retrieve public key
	router.POST("/conceptindexor/concept", createConcept)   // hash a concept (and save the salt if needed)
	router.DELETE("/conceptindexor/concept", deleteConcept) // delete a salt in the database

}
