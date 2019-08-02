package subscribe

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cozy/cozy-stack/pkg/dispers"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/dispers/subscribe"
	"github.com/cozy/echo"
)

/*
*
*
TARGET FINDER'S ROUTES : those functions are used on route ./dispers/targetfinder/
*
*
*/

func decryptList(c echo.Context) error {

	var in subscribe.InputDecrypt
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	// TODO: Decrypt inputs / decrypt Instances and Encrypt for T

	return c.JSON(http.StatusOK, subscribe.InputInsert{
		IsEncrypted:        in.IsEncrypted,
		EncryptedInstances: in.EncryptedInstances,
		EncryptedInstance:  in.EncryptedInstance,
	})
}

func encryptList(c echo.Context) error {

	var in subscribe.InputEncrypt
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	// TODO : Decrypt Inputs and Encrypt for TF

	return c.JSON(http.StatusOK, subscribe.OutputEncrypt{
		IsEncrypted:        in.IsEncrypted,
		EncryptedInstances: in.EncryptedInstances,
	})
}

/*
*
*
Target's ROUTES : those functions are used on route ./dispers/target/
*
*
*/

func insert(c echo.Context) error {

	var in subscribe.InputInsert
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}
	// TODO : Decrypt inputs before unmarshalling

	var listOfInstances []query.Instance
	if in.EncryptedInstances != nil {
		err := json.Unmarshal(in.EncryptedInstances, &listOfInstances)
		if err != nil {
			return err
		}
	}
	var instance query.Instance
	err := json.Unmarshal(in.EncryptedInstance, &instance)
	if err != nil {
		return err
	}

	// Search if instance is already present in list
	absent := true
	for index := 0; absent && index < len(listOfInstances); index++ {
		if instance.Domain == listOfInstances[index].Domain {
			instance.Version = listOfInstances[index].Version + 1
			listOfInstances[index] = instance
			absent = false
		}
	}
	if absent {
		instance.Version = 1
		listOfInstances = append(listOfInstances, instance)
	}

	if len(instance.TokenBearer) < 5 {
		return errors.New("Unvalid token bearer")
	}

	encListOfInstances, err := json.Marshal(listOfInstances)
	if err != nil {
		return nil
	}

	return c.JSON(http.StatusOK, subscribe.InputEncrypt{
		IsEncrypted:        in.IsEncrypted,
		EncryptedInstances: encListOfInstances,
	})
}

/*
*
*
Conductor's ROUTES : those functions are used on route ./dispers/conductor
*
*
*/

func createConceptInConductorDB(c echo.Context) error {

	// Unmarshal input
	var in query.InputCI
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	if err := enclave.CreateConceptInConductorDB(&in); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ok": true,
	})
}

func subscribeToRequest(c echo.Context) error {

	var in subscribe.InputConductor
	if err := json.NewDecoder(c.Request().Body).Decode(&in); err != nil {
		return err
	}

	if !in.IsEncrypted {
		encInst, err := json.Marshal(in.Instance)
		if err != nil {
			return err
		}
		in.EncryptedInstance = encInst
	}

	if err := enclave.Subscribe(&in); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ok": true,
	})
}

// Routes sets the routing for the dispers service
func Routes(router *echo.Group) {

	router.POST("/targetfinder/decrypt", decryptList)
	router.POST("/targetfinder/encrypt", encryptList)

	router.POST("/target/insert", insert)

	router.POST("/conductor/concept", createConceptInConductorDB)
	router.POST("/conductor/subscribe", subscribeToRequest)
}
