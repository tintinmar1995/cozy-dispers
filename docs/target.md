# Cozy-DISPERS Target 

1. How-To
- How to get the queries ?
- How to add an instance to a list ? 
2. Serialization of every information
3. Functions and tests

## How-To

### How to get the queries ?

**Step 1 :** Create the local query

To get a query for each target, you need to give Target the Targets and the Local Query. Create, then, an InputT object. To do that, you will need to create a Local Query and pass the array of instances retrieved from TF or from the conductor's database.

```golang
type LocalQuery struct {
	Doctypes    []string  `json:"doctypes,omitempty"`
	IndexFields []string  `json:"indexfield,omitempty"`
	Filters     mango.Map `json:"filters,omitempty"`
}
```
A local query possibly requires several doctypes. Mango queries define a way to select information regarding selector. 

NB : To see examples of Filters, [read mango file's test](https://github.com/tintinmar1995/cozy-stack/blob/enclave-tf/pkg/couchdb/mango/query_test.go)

**Step 2 :** Pass the list of targets

```golang
type Instance struct {
	Host   string `json:"host,omitempty"`
	Domain string `json:"domain,omitempty"`
	Token  Token  `json:"token,omitempty"`
}
```
To make its treatment, Target has access to information about an instance : 
- Host's URL (e.g. `mycozy.cloud`)
- Domain (e.g. `prettyname4acozy`)
- A token given by the instance during the subscription process

**Step 3 :** Create the input structure and make the request

```golang
type InputT struct {
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	Targets             []Instance `json:"Addresses,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTargets    []byte     `json:"enc_addresses,omitempty"`
}
```

NB : In a proper DISPERS, data will be encrypted. 

```http
POST /target/gettokens HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputT
```

**Step 4 :** get the result

You will receive an array of queries. Each query sums up every information needed to query a stack. Host has to be read by the Conductor to address to query to the right stack. 

```golang
type Query struct {
	Host                string     `json:"host,omitempty"`
	Domain              string     `json:"domain,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	TokenBearer         string     `json:"bearer,omitempty"`
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	EncryptedDomain     []byte     `json:"enc_domain,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTokens     []byte     `json:"enc_token,omitempty"`
}
```

### How to add an instance to a list ? 

.. TODO

## Functions and tests

