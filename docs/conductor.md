# Cozy-DISPERS : Conductor

Conductor leads the Query, pass the right information to the right enclave. Almost every treatment made by the conductor could be made by you ... except one special thing  

1. How-to retrieve the list of instances associated to a concept
2. How-to add an instance to Conductor's Database
3. Conductor's Database

## How-to retrieve the list of instances associated to a concept

```golang
func RetrieveSubscribeDoc(hash string) ([]SubscribeDoc, error) {

	var out []SubscribeDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("hash", hash)}
	err = couchdb.FindDocs(prefixerC, "io.cozy.shared4ml", req, &out)
	if err != nil {
		return out, err
	}

	if len(out) > 1 {
		return out, errors.New("There is more than 1 subscribe doc in database for this concept")
	}

	return out, nil
}
```

## How-to add an instance to Conductor's Database

[See the process here](subscribe.md)

## Conductor's Database

```golang
type SubscribeDoc struct {
	SubscribeID        string `json:"_id,omitempty"`
	SubscribeRev       string `json:"_rev,omitempty"`
	Hash               string `json:"hash,omitempty"`
	EncryptedInstances []byte `json:"enc_instances"`
}
```

| Index  | Actual Value  | Conductor's point of view  | TF's point of view  |
| ------ | ------------------------: | ------------------------: | ------------------------: |
| hash1  | *list of instances* | *encrypted information* | *list of encrypted information* |
| hash2  | [inst1, inst2, ...] | 7kfRLc    | [QYcTLi, f3YZBW, ...] |
| hash3  | ... |   ...     | ... |


```golang
type Token struct {
	TokenBearer string `json:"bearer,omitempty"`
}

type Instance struct {
	Domain           string    `json:"domain"`
	SubscriptionDate time.Time `json:"date"`
	Token            Token     `json:"token"`
}
```
