# Cozy-DISPERS : Concept Indexor

1. How-To
2. Unexported functions
3. Exported functions
4. Tests

## How-To

### Get one concept's hash

In test mode, you can choose to encrypt or not the concept. In a proper Cozy-DISPERS request, the concept will be encrypt and it will be impossible to choose.

**Step 1** : Make the request

To get one concept's hash, you will need to make a GET request. You can hash several concepts in a single request.

At the end of the URL, you will have to choose if it is an encrypted request or not.

```http
GET query/conceptindexor/concept/concept1:concept2:concept3/false HTTP/1.1
Host: cozy.example.org
Content-Type: application/json
```

**Step 2** : Get the result

You will then received an answer that fit this structure.

```golang
type OutputCI struct {
	Hashes []Concept `json:"hashes,omitempty"`
}
```

### Add a concept to the database

**Step 1** : Create a Concept

In test mode, you can choose to encrypt or not the concept. In a proper Cozy-DISPERS request, the concept will be encrypt and it will be impossible to choose.

```golang
type Concept struct {
	IsEncrypted		bool 	 `json:"is_encrypted,omitempty"`
	Concept       string `json:"concept,omitempty"`
}
```

**Step 2** : Create an input

Once the concept created, for example `Concept {IsEncrypted : false, Concept : "Cozy-Cloud"}`. Create an InputCI, and indicate if the concept is encrypted.

```golang
type InputCI struct {
	Concepts    []Concept `json:"concepts,omitempty"`
}
```

**Step 3** : Make the request

```http
POST query/conceptindexor/concept HTTP/1.1
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

### Delete a concept from the database

**Step 1** : Make the request

The structure to delete a concept is quite like the get one.

```http
DELETE query/conceptindexor/concept/concept1:concept3/false HTTP/1.1
Host: cozy.example.org
Content-Type: application/json
```

## Unexported functions

1. `createHash` : generate one hash from one concept and a randomly generated salt. Salts are random bytes of size 128.

2. `getHash` : retrieve one hash from the database with the concept. This function uses mango query and answer an error if there is no existing doc or if there are more than one doc.

3. `saveConcept` : create a hash with createHash and save it in conceptindexor's database

4. `isConceptExisting` : this function is only used to test if the other functions work properly.

## Exported functions

1. `CreateConcept` : used to create a concept's hash from a decrypted concept

2. `GetConcept` : used to get a concept's hash from a decrypted concept

3. `DecryptConcept` .. TODO

4. `DeleteConcept` : used to delete a concept's hash from a decrypted concept

## Tests

1. Test the function isConceptExisting
2. Test if a new concept exists (concept randomly generated)
3. Create twice the same salt and check if there is error when getting salt
4. Test if hash generation is not constant
5. Test consistency when asking twice the same concept
