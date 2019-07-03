package dispers

import (
	"encoding/json"
	"net/http"

	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/cozy/echo"
)

/*
*
*
Target'S ROUTES : those functions are used on route ./dispers/target/
*
*
*/
func getData(c echo.Context) error {

	var localQuery dispers.InputT

	if err := json.NewDecoder(c.Request().Body).Decode(&localQuery); err != nil {
		return err
	}

	data, err := enclave.GetData(localQuery)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dispers.OutputT{Data: data})
}

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {

	router.POST("/target/getdata", getData)

}
