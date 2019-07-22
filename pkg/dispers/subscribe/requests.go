package subscribe

import "github.com/cozy/cozy-stack/pkg/couchdb"

// SubscribeDoc is used to save in Conductor's database the list of instances that subscribed
// to a concept. Each concept is associated with a list of instances.
type SubscribeDoc struct {
	SubscribeID        string `json:"_id,omitempty"`
	SubscribeRev       string `json:"_rev,omitempty"`
	Hash               []byte `json:"hash"`
	EncryptedInstances []byte `json:"enc_instances"`
}

// ID is used to get SubscribeID
func (t *SubscribeDoc) ID() string {
	return t.SubscribeID
}

// Rev is used to get SubscribeRev
func (t *SubscribeDoc) Rev() string {
	return t.SubscribeRev
}

// DocType is used to get the doc's type
func (t *SubscribeDoc) DocType() string {
	return "io.cozy.instances"
}

// Clone is used to copy one doc
func (t *SubscribeDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

// SetID is used to set doc's ID
func (t *SubscribeDoc) SetID(id string) {
	t.SubscribeID = id
}

// SetRev is used to set doc's Rev
func (t *SubscribeDoc) SetRev(rev string) {
	t.SubscribeRev = rev
}

type InputConductor struct {
	IsEncrypted       bool     `json:"is_encrypted,omitempty"`
	Concepts          []string `json:"concepts,omitempty"`
	EncryptedInstance []byte   `json:"enc_instance,omitempty"`
}

type InputDecrypt struct {
	IsEncrypted        bool   `json:"is_enc"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
	EncryptedInstance  []byte `json:"enc_instance"`
}

type InputInsert struct {
	IsEncrypted        bool   `json:"is_enc"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
	EncryptedInstance  []byte `json:"enc_instance"`
}

type InputEncrypt struct {
	IsEncrypted        bool   `json:"is_enc,omitempty"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
}

type OutputEncrypt struct {
	IsEncrypted        bool   `json:"is_enc,omitempty"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
}
