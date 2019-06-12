package dispers

/*
*
Queries' Input & Output
*
*/

/*
*
Concept Indexors' Input & Output
*
*/

type Concept struct {
	Concept          string `json:"concept,omitempty"`
	EncryptedConcept []byte `json:"enc_concept,omitempty"`
	Hash             string `json:"hash,omitempty"`
}

type InputCI struct {
	Concepts    []Concept `json:"concepts,omitempty"`
	IsEncrypted bool      `json:"encrypted,omitempty"`
}

// OutputCI contains a bool and the result
type OutputCI struct {
	Hashes []Concept `json:"hashes,omitempty"`
}
