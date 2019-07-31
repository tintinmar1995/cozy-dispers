package enclave

import (
	"errors"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func decryptInputsTF(in *query.InputTF) error {

	if in.IsEncrypted {
		// TODO: decrypt input byte and save it in in.ListsOfAddresses
	}

	return nil
}

// SelectAddresses apply the target profile over lists of addresses
func SelectAddresses(in query.InputTF) ([]string, error) {

	if err := decryptInputsTF(&in); err != nil {
		return nil, err
	}

	finalList, err := in.TargetProfile.Compute(in.ListsOfAddresses)
	if len(finalList) == 0 {
		return nil, errors.New("Result of target profile is empty")
	}
	// TODO: Encrypt final list
	return finalList, err

}
