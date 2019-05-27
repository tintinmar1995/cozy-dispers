package dispers

import (
	"net/http"

	"github.com/cozy/echo"
)

/*
*
*
COMMON ROUTES : those 3 functions are used on route ./dispers/
*
*
*/
func index(c echo.Context) error {

	out := "Hello ! You reach correctly the learning part of the Cozy Learning Server."
	return c.String(http.StatusOK, out)
}

func indexBis(c echo.Context) error {
	out := ""
	actor := c.Param("actor")
	switch actor {
	case "conductor":
		out = "Hello ! I'm ruling the game !"
	case "conceptindexor":
		out = "Hello ! I will do Concept Indexor's dishes !"
	case "_target":
		out = "Hello ! I will do Target's dishes !"
	case "dataaggregator":
		out = "Hello ! I will do Data Aggregator's dishes !"
	case "targetfinder":
		out = "Hello ! I will do Target Finder's dishes !"
	default:
		return nil
	}

	return c.String(http.StatusOK, out)
}

func getPublicKey(c echo.Context) error {
	out := ""
	actor := c.Param("actor")
	switch actor {
	case "conductor":
		out = "DxXa9Bqe7Fb5G4MgZ6dmXAr7v33mIuY9X"
	case "conceptindexor":
		out = "tboHCSnIPcvvX9BI89yKKGKp4u2Ra3zsP"
	case "target":
		out = "wCYxkY2RUgfisQDQnwN7yi9ur3gdKk782"
	case "dataaggregator":
		out = "X9IE7UQ4ZfXQ5jRsPbeJHRsWy4WZSwnjk"
	case "targetfinder":
		out = "Psy8PB5o6WL3PkccoLrF4pSfpr2dDPaxe"
	default:
		return nil
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ok":  true,
		"key": out,
	})
}

/*
*
*
CONCEPT INDEXOR'S ROUTES : those functions are used on route ./dispers/conceptindexor/
*
*
*/
func allConcepts(c echo.Context) error {
	list, err := enclave.GetAllConcepts()
	return c.JSON(http.StatusCreated, echo.Map{
		"ok":       err == nil,
		"concepts": list,
	})
}

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
	// API's Index
	router.GET("/", index)
	router.GET("/:actor", indexBis)
	router.GET("/:actor/publickey", getPublicKey)

	router.GET("/conceptindexor/allconcepts", allConcepts)
	router.POST("/conceptindexor/hash/concept=:concept", hashConcept)     // hash a concept (and save the salt if needed)
	router.DELETE("/conceptindexor/hash/concept=:concept", deleteConcept) // delete a salt in the database

}
