package dispers

import "time"

/*
*
Targets' Input & Output
*
*/

// InputT contains information received by Target's enclave
type InputT struct {
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	Targets             []Instance `json:"Addresses,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTargets    []byte     `json:"enc_addresses,omitempty"`
}

// Query is all the information need by the conductor's and stack to make a query
type Query struct {
	Domain              string     `json:"domain,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	TokenBearer         string     `json:"bearer,omitempty"`
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTokens     []byte     `json:"enc_token,omitempty"`
}

// OutputT is what Target returns to the conductor
type OutputT struct {
	Data []map[string]interface{} `json:"data,omitempty"` // type Query
}

// LocalQuery decribes which data the stack has to retrieve
type LocalQuery struct {
	FindRequest map[string]interface{} `json:"findrequest,omitempty"`
}

// Token is used to serialize the token
type Token struct {
	TokenBearer string `json:"bearer,omitempty"`
}

// Instance describes the location of an instance and the token it had created
// When Target received twice the same Instance, it needs to be able to consider the more recent item
type Instance struct {
	Domain           string    `json:"domain"`
	SubscriptionDate time.Time `json:"date"`
	Token            Token     `json:"token"`
}
