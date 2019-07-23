# Cozy-DISPERS Target

1. How-To retrieve data ?

## How-To

## How-to retrieve data ?

**Step 1 :** Create the local query

To get a query for each target, you need to give Target the Targets and the Local Query. Create, then, an InputT object. To do that, you will need to create a Local Query and pass the array of instances retrieved from TF or from the conductor's database.

```golang
type LocalQuery struct {
	FindRequest map[string]interface{} `json:"findrequest,omitempty"`
}
```

FindRequest follows the [CouchDB syntax](https://docs.couchdb.org/en/stable/api/database/find.html?highlight=selector) for queries.

**Step 2 :** Pass the list of targets

```golang
type Instance struct {
	Domain           string    `json:"domain"`
	SubscriptionDate time.Time `json:"date"`
	Token            Token     `json:"token"`
}
```

To make its treatment, Target has access to information about an instance :
- Domain (e.g. `prettyname4acozy.mycozy.cloud`)
- A token given by the instance during the subscription process

**Step 3 :** Create the input structure and make the request

```golang
type InputT struct {
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	Targets             []string   `json:"Addresses,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTargets    []byte     `json:"enc_addresses,omitempty"`
}
```

NB : Behind []string Targets, we found a list of marshalled Instances

```http
POST query/target/query HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputT
```

**Step 4 :** get the result

Target is creating an array of queries. Each query sums up every information needed to query a stack. Host has to be read by the Conductor to address to query to the right stack.

```golang
type Query struct {
	Domain              string     `json:"domain,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	TokenBearer         string     `json:"bearer,omitempty"`
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTokens     []byte     `json:"enc_token,omitempty"`
}
```

From each query, Target is making a HTTP Request to stacks where instances are hosted. Target uses the route `find`

Finally, Target gathers all those documents in a `[]map[string]interface{}` structure and passes it to the conductor.
