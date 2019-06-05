# Cozy-DISPERS : Concept Indexor

Even if there is several Cozy-DISPERS server hosted, one server's conductor will always use the same server to decrypt a concept and get the associated hash. It prevents from dealing with consistency between one concept's hashes.

## How-To

### Get one concept's hash

**Step 1** : Create a Concept

You can choose to encrypt or not the concept. In a proper Cozy-DISPERS request, the concept will be encrypt and it will be impossible to choose.

```golang
type Concept struct {
	Concept          string `json:"concept,omitempty"`
	EncryptedConcept []byte `json:"enc_concept,omitempty"`
	Hash             string `json:"hash,omitempty"`
}
```

**Step 2** : Create an input

Once the concept created, for example `Concept {	Concept : "Cozy-Cloud"}`. Create an InputCI, and indicate if the concept is encrypted.

```golang
type InputCI struct {
	Concepts    []Concept `json:"concepts,omitempty"`
	IsEncrypted bool      `json:"encrypted,omitempty"`
}
```

**Step 3** : Make the request

```http
GET /conceptindexor/concept HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputCI
```

### Add a concept to the database

Simply ask the API to get the concept's hash and it will be automatically created and saved in the database.

### Delete a concept from the database

## Test
