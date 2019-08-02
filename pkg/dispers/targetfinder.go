package enclave

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func decryptInputsTF(in *query.InputTF) error {

	if in.IsEncrypted {
		// TODO: decrypt input byte and save it in in.ListsOfAddresses
	}

	return nil
}

func developpTargetProfile(compressedTP string) (string, error) {

	compressedTP = strings.ReplaceAll(compressedTP, "::", "},\"right_node\": {\"type\": 0,\"value\": ")
	compressedTP = strings.ReplaceAll(compressedTP, "(\"", "({\"type\": 0,\"value\": \"")
	compressedTP = strings.ReplaceAll(compressedTP, "\")", "\"}")
	compressedTP = strings.ReplaceAll(compressedTP, ":OR(", "},\"right_node\": OR(")
	compressedTP = strings.ReplaceAll(compressedTP, ":AND(", "},\"right_node\": AND(")
	compressedTP = strings.ReplaceAll(compressedTP, "OR(", "{\"type\": 1,\"left_node\": ")
	compressedTP = strings.ReplaceAll(compressedTP, "AND(", "{\"type\": 2,\"left_node\": ")
	compressedTP = strings.ReplaceAll(compressedTP, ")", "}")
	compressedTP = compressedTP + "}"

	return compressedTP, nil
}

// SelectAddresses apply the target profile over lists of addresses
func SelectAddresses(in query.InputTF) ([]string, error) {

	if err := decryptInputsTF(&in); err != nil {
		return nil, err
	}

	// Translate string target profile to OperationTree
	jsonTP, err := developpTargetProfile(in.TargetProfile)
	if err != nil {
		return nil, err
	}
	var targetProfile query.OperationTree
	if err := json.Unmarshal([]byte(jsonTP), &targetProfile); err != nil {
		return nil, err
	}

	finalList, err := targetProfile.Compute(in.ListsOfAddresses)
	if len(finalList) == 0 {
		return nil, errors.New("Result of target profile is empty")
	}
	// TODO: Encrypt final list
	return finalList, err

}
