package errors

import (
	"errors"

	"github.com/cozy/cozy-stack/pkg/jsonapi"
)

var (

	// Global
	ErrUnmarshal           = errors.New("Failed to unmarshal json")
	ErrRouteNotFound       = errors.New("Unknown route")
	ErrAsyncTypeUnknown    = errors.New("Unknown AsyncType")
	ErrAsyncTaskNotFound   = errors.New("Async task not found")
	ErrTooManyDoc          = errors.New("One unique doc expected, but more than one found")
	ErrNoExecutionMetadata = errors.New("No ExecutionMetadata for this query")

	// CI
	ErrEmptyConcept           = errors.New("Concept is empty")
	ErrConceptNotFound        = errors.New("Conceptdoc not found")
	ErrConceptAlreadyExisting = errors.New("Concept is already existing")
	ErrTooManyConceptDoc      = errors.New("Too many salts found for this concept")

	// TF / T
	ErrInvalidTargetProfile = errors.New("Invalid target profile")
	ErrComputeTargetProfile = errors.New("Failed to compute target profile")
	ErrNoTargets            = errors.New("No target matches the Target Profile")

	// DA
	ErrArgNotFound       = errors.New("Arg not found")
	ErrKeyNotFound       = errors.New("Key not found")
	ErrInvalidKey        = errors.New("Invalid key")
	ErrStrToFloat        = errors.New("Cannot convert string to float64")
	ErrAggrUnknown       = errors.New("Unknown aggregation function")
	ErrAggrFailed        = errors.New("Failed to apply aggregate function")
	ErrLengthConsistency = errors.New("Theta and features should have the same length")

	// Conductor
	ErrHostnameConductor           = errors.New("Failed to retrieve hostname")
	ErrNewExecutionMetadata        = errors.New("Failed to create a new ExecutionMetadata")
	ErrRetrievingQueryDoc          = errors.New("Failed while retrieving QueryDoc")
	ErrRetrievingExecutionMetadata = errors.New("Failed while retrieving ExecutionMetadata")
	ErrSubscribeDocNotFound        = errors.New("Cannot find SubscribeDoc")
	ErrNotEnoughDataToComputeQuery = errors.New("We don't have enough data to compute the query")
	ErrConceptAlreadyInConductorDB = errors.New("This concept already exists in Conductor's database")
)

func WrapErrors(err error, parameter string) error {

	if err == nil {
		return nil
	}

	switch err {
	case ErrRouteNotFound:
		return jsonapi.NotFound(err)
	case ErrAsyncTaskNotFound:
		return jsonapi.NotFound(err)
	case ErrConceptNotFound:
		return jsonapi.NotFound(err)
	case ErrArgNotFound:
		return jsonapi.NotFound(err)
	case ErrKeyNotFound:
		return jsonapi.NotFound(err)
	case ErrSubscribeDocNotFound:
		return jsonapi.NotFound(err)
	case ErrNoTargets:
		return jsonapi.NotFound(err)
	case ErrNoExecutionMetadata:
		return jsonapi.NotFound(err)
	case ErrEmptyConcept:
		return jsonapi.InvalidParameter(parameter, err)
	case ErrAggrUnknown:
		return jsonapi.InvalidParameter(parameter, err)
	case ErrInvalidTargetProfile:
		return jsonapi.InvalidParameter(parameter, err)
	case ErrInvalidKey:
		return jsonapi.InvalidParameter(parameter, err)
	case ErrUnmarshal:
		return jsonapi.BadJSON()
	case ErrNotEnoughDataToComputeQuery:
		return jsonapi.Forbidden(err)
	default:
		return jsonapi.InternalServerError(err)
	}

}
