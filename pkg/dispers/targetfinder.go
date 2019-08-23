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

	compressedTP = strings.ReplaceAll(compressedTP, ")", ":):")
	secondSplit := []string{}
	firstSplit := strings.Split(compressedTP, "(")
	for _, a := range firstSplit {
		secondSplit = append(secondSplit, strings.Split(a, ":")...)
	}

	// Remove empty element
	index := 0
	for index < len(secondSplit) {
		if len(secondSplit[index]) < 2 && secondSplit[index] != ")" {
			if index != len(secondSplit)-1 {
				secondSplit = append(secondSplit[:index], secondSplit[index+1:]...)
			} else {
				secondSplit = secondSplit[:index]
			}
		} else {
			index = index + 1
		}
	}

	// Look for leafs and wrap them with json
	for index, item := range secondSplit {
		if item != "OR" && item != "AND" && item != ")" {
			secondSplit[index] = "{\"type\":0,\"value\":\"" + item + "\"}"
		}
	}

	// Look for ) and aggregate OR/AND node 1 and node 2
	var op string
	var nodeA string
	var nodeB string
	index = 0
	jsonNode := "{\"type\"://TYPE//,\"left_node\"://LEFT//,\"right_node\"://RIGHT//}"
	for index < len(secondSplit) {
		if secondSplit[index] == ")" {
			op = secondSplit[index-3]
			nodeA = secondSplit[index-2]
			nodeB = secondSplit[index-1]

			// remove nodeA
			secondSplit = append(secondSplit[:index-2], secondSplit[index-2+1:]...)
			// remove nodeB
			secondSplit = append(secondSplit[:index-2], secondSplit[index-2+1:]...)
			// remove )
			if index-2 != len(secondSplit)-1 {
				secondSplit = append(secondSplit[:index-2], secondSplit[index-2+1:]...)
			} else {
				secondSplit = secondSplit[:index-2]
			}

			// aggregate to build json and replace with op
			if op == "AND" {
				op = strings.ReplaceAll(jsonNode, "//TYPE//", "2")
			} else if op == "OR" {
				op = strings.ReplaceAll(jsonNode, "//TYPE//", "1")
			}
			op = strings.ReplaceAll(op, "//LEFT//", nodeA)
			op = strings.ReplaceAll(op, "//RIGHT//", nodeB)
			secondSplit[index-3] = op
			index = 0
		} else {
			index = index + 1
		}
	}

	return secondSplit[0], nil
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
