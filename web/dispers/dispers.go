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

func hashConcept(c echo.Context) error {
	concept := c.Param("concept")
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
	router.POST("/conceptindexor/hash/concept=:concept", hashConcept)     // hash a concept (and save the salt if needed)
	router.DELETE("/conceptindexor/hash/concept=:concept", deleteConcept) // delete a salt in the database

}
