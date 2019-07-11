package enclave

import (
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
		return []string{}, err
	}

	finalList, err := in.TargetProfile.Compute(in.ListsOfAddresses)
	// TODO: Encrypt final list
	return finalList, err

}
