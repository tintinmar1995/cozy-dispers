# Cozy-DISPERS : Concept Indexor

Even if there is several Cozy-DISPERS server hosted, one server's conductor will always use the same server to decrypt a concept and get the associated hash. It prevents from dealing with consistency between one concept's hashes.

1. How-To
2. Unexported functions
3. Exported functions
4. Tests

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

**Step 4** : Get the result

You will then received an answer of this structure.

```golang
type OutputCI struct {
	Hashes []Concept `json:"hashes,omitempty"`
}
```

### Add a concept to the database

Simply ask the API to get the concept's hash and it will be automatically created and saved in the database.

### Delete a concept from the database

**Step 1** : Create a Concept

Same as creating a concept, choose between encrypted/decrypted mode.

```golang
type Concept struct {
	Concept          string `json:"concept,omitempty"`
	EncryptedConcept []byte `json:"enc_concept,omitempty"`
	Hash             string `json:"hash,omitempty"`
}
```

NB : Specify a hash will be completely useless

**Step 2** : Make the request

```http
DELETE /conceptindexor/concept HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputCI
```

## Unexported functions

1. `addSalt` : generate and add a salt in the database. The function does not check if one salt is already existing. Salts are random bytes of size 128. Hashes are made with scrypt algorithm using those parameters and saved with the following doc :  

```golang
ConceptDoc{
		ConceptID:  "",
		ConceptRev: "",
		Concept:    concept,
		Salt:       string(crypto.GenerateRandomBytes(defaultLenSalt)),
}
```

2. `getSalt` : retrieve one salt from the database with the concept. This function uses mango query and answer an error if there is no existing salt or if there are more than one salts.

3. `hash` : Hashes are made using scrypt algorithm with those parameters :

```golang
const (
	defaultN       = 32768
	defaultR       = 8
	defaultP       = 1
	defaultDkLen   = 32 // hash length
)
```

4. `isConceptExisting` : this function is only used to test if the other functions work properly.

## Exported functions

1. `CreateConcept` : used to get a concept's hash from a decrypted concept

2. `DecryptConcept` .. TODO

3. `DeleteConcept` : used to delete a concept's hash from a decrypted concept

## Tests

1. Test the function isConceptExisting
2. Test if a new concept exists (concept randomly generated)
3. Create twice the same salt and check if there is error when getting salt
4. Test if hash generation is not constant
5. Test consistency when asking twice the same concept
