# Cozy-DISPERS Subscription Process

1. How-to add an instance to Cozy-DISPERS
2. How-to add a concept to Cozy-DISPERS
3. Steps of the process

## How-to add an instance to Cozy-DISPERS

```golang
type InputConductor struct {
	IsEncrypted       bool     `json:"is_encrypted,omitempty"`
	Concepts          []string `json:"concepts,omitempty"`
	EncryptedInstance []byte   `json:"enc_instance,omitempty"`
}
```

```http
POST subscribe/conductor/subscribe HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputConductor
```

## How-to add a concept to Cozy-DISPERS

```golang
type InputCI struct {
	IsEncrypted       bool      `json:"is_encrypted,omitempty"`
	Concepts          []Concept `json:"concepts,omitempty"`
}
```

```http
POST subscribe/conductor/concept HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputCI
```


## Steps of the process

**Step 1 :** Get hashes

```http
GET query/conceptindexor/concept1-concept2/true HTTP/1.1
Host: cozy.example.org
Content-Type: application/json
```

**Step 2 :** Retrieve lists of instances from Conductor's database

**Step 3 :** Decrypt lists of instances for Target

```golang
type InputDecrypt struct {
	IsEncrypted        bool
	EncryptedInstances []byte
}
```

```http
POST subscribe/targetfinder/decrypt HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputDecrypt
```

**Step 4 :** Append the new instances

```golang
type InputInsert struct {
	IsEncrypted        bool   `json:"is_enc,omitempty"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
	EncryptedInstance  []byte `json:"enc_instance,omitempty"`
}
```

```http
POST subscribe/target/insert HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputInsert
```

**Step 5 :** Encrypt lists of instances and save the new lists

```golang
type InputEncrypt struct {
	IsEncrypted        bool   `json:"is_enc,omitempty"`
	EncryptedInstances []byte `json:"enc_instances,omitempty"`
}
```

```http
POST subscribe/targetfinder/encrypt HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputEncrypt
```
