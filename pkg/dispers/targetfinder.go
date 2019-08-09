package enclave

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
)

func decryptInputsTF(in *query.InputTF) (string, map[string][]string, error) {

	if in.IsEncrypted {
		// TODO: decrypt input byte and save it in in.ListsOfAddresses
	}

	compressedTP := string(in.EncryptedTargetProfile)
	if len(compressedTP) < 4 {
		return "", nil, errors.New("Invalid target profile")
	}

	var listOfAddresses []string
	listsOfAddresses := make(map[string][]string)
	for key, encryptedList := range in.EncryptedListsOfAddresses {
		if in.IsEncrypted {
			listOfAddresses = []string{}
			for _, addresse := range strings.Split(string(encryptedList)[2:len(encryptedList)-2], ",") {
				listOfAddresses = append(listOfAddresses, addresse)
			}
			listsOfAddresses[key] = listOfAddresses
		} else {
			listOfAddresses = []string{}
			// for tests only
			strEncryptedList := strings.ReplaceAll(string(encryptedList), "}\",\"{", "},{")
			strEncryptedList = strings.ReplaceAll(strEncryptedList, "[\"{", "[{")
			strEncryptedList = strings.ReplaceAll(strEncryptedList, "}\"]", "}]")

			for _, addresse := range strings.Split(strEncryptedList[2:len(strEncryptedList)-2], "},{") {
				listOfAddresses = append(listOfAddresses, "{"+addresse+"}")
			}
			listsOfAddresses[key] = listOfAddresses
		}
	}

	return compressedTP, listsOfAddresses, nil
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

	compressedTP, listsOfAddresses, err := decryptInputsTF(&in)
	if err != nil {
		return nil, errors.New("Cannot decrypt concept : " + err.Error())
	}

	// Translate string target profile to OperationTree
	jsonTP, err := developpTargetProfile(compressedTP)
	if err != nil {
		return nil, err
	}
	var targetProfile query.OperationTree
	if err := json.Unmarshal([]byte(jsonTP), &targetProfile); err != nil {
		return nil, errors.New("Failed to unmarshal Target Profile : " + err.Error())
	}

	finalList, err := targetProfile.Compute(listsOfAddresses)
	if err != nil {
		return nil, errors.New("Failed to compute target profile : " + err.Error())
	}
	if len(finalList) == 0 {
		return nil, errors.New("Result of target profile is empty")
	}
	// TODO: Encrypt final list

	return finalList, nil

}
