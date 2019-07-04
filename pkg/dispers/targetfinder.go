package enclave

import (
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
)

func decryptInputsTF(in *dispers.InputTF) error {

	if in.IsEncrypted {
		// TODO: decrypt input byte and save it in in.ListsOfAddresses
	}

	return nil
}

// SelectAddresses apply the target profile over lists of addresses
func SelectAddresses(in dispers.InputTF) ([]string, error) {

	if err := decryptInputsTF(&in); err != nil {
		return []string{}, err
	}

	finalList, err := in.TargetProfile.Compute(in.ListsOfAddresses)
	// TODO: Encrypt final list
	return finalList, err

}

// CreateTokens is used by the conductor after a client subscribe to data sharing
func CreateTokens(inst string, listOfInstances []string) ([]string, error) {

	// TODO : Decrypt list
	listOfInstances = append(listOfInstances, inst)
	// TODO: Encrypt list

	return listOfInstances, nil
}
