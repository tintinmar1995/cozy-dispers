/*
In this file, we lead a DISPERS-ML learning. We're going to choose the actors in
the process and talk with them. We're probabliy going to interact with several
stacks and several servers. This script is also going to keep the querier (front
-end) acknowledge of the process by updating repeatedly the doc in his Couchdb.
A conductor is instanciated when a user call the associated route of this API.
*/

package enclave

import (
    //"strings"
    //"encoding/json"
    //"io/ioutil"
  	//"net/http"

  	"github.com/cozy/echo"
    "github.com/cozy/cozy-stack/pkg/couchdb"
    "github.com/cozy/cozy-stack/pkg/prefixer"
    "github.com/cozy/cozy-stack/pkg/dispers/dispers"
)

/*
Doc used to saved in Conductor's database the list of instances that subscribed
to a concept
*/
type SubscribeDoc struct {
	SubscribeID  string `json:"_id,omitempty"`
	SubscribeRev string `json:"_rev,omitempty"`
	Adresses     string `json:"adresses"`
}

func (t *SubscribeDoc) ID() string {
	return t.SubscribeID
}

func (t *SubscribeDoc) Rev() string {
	return t.SubscribeRev
}

func (t *SubscribeDoc) DocType() string {
	return "io.cozy.shared4ml"
}

func (t *SubscribeDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

func (t *SubscribeDoc) SetID(id string) {
	t.SubscribeID = id
}

func (t *SubscribeDoc) SetRev(rev string) {
	t.SubscribeRev = rev
}


// Subscribe is used by a user to share a new data
func Subscribe(domain, prefix string, adresses []string){
  /*
  couchdb.EnsureDBExist(prefixer.ConductorPrefixer, "io.cozy.shared4ml")

  doc := &SubscribeDoc{
    SubscribeID: "subscription",
		Adresses: adresses,
	}

  couchdb.CreateNamedDocWithDB(mPrefixer, doc)
  // TO DO : update doc in conductor/shared4ml
  */
}


/*
The structure actor gives the conductor a way to consider every distant actors
and to communicate with it.
*/
type actor struct{
  host      string
  api       string
  out       map[string]interface{}
  outmeta   []actor
}

/*
func (a actor) makeRequestGet(job string) error {

  url := strings.Join([]string{"http://", a.host, "/", a.api,"/", job}, "")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.NewDecoder(body).Decode(&a.out)
	return nil

}

func MakeRequestPost(job string, data string) map[string]interface{} {

  url := strings.Join([]string{"http://", a.host, "/", a.api,"/", job}, "application/json", bytes.NewBuffer(data))

	resp, err := http.Post(url, )
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

  var result map[string]interface{}
  json.NewDecoder(body).Decode(&a.out)
	return nil

}
*/

// makeRequestPatch

// makeRequestDelete

/*
The script is going to retrieve informations in the querier's db and follows
this informations to the different api. The most part of this informations is
encrypted, the conductor is not supposed to deduced anything from all what he is
manipulating.
*/
type Training struct {
  AlgoML         string   `json:"algo,omitempty"` // model trained
  Dataset        string   `json:"dataset,omitempty"` // dataset used
  DispersAlgo    string   `json:"dispersalgo,omitempty"` // typo from dispers-ml
  FormulaTarget  string   `json:"formulatarget,omitempty"` // var from dataset to predict
  FormulaPreds   []string `json:"formulapredictors,omitempty"` // predictors from dataset
  State          string   `json:"state,omitempty"`
  // parameters
}

/*
Every training has metadata saved in the conductor's database. Thanks to that,
the querier can retrieve the learning's state.
*/
type queryDoc struct {
	queryID     string         `json:"_id,omitempty"`
	queryRev    string         `json:"_rev,omitempty"`
  MyTraining  Training       `json:"training,omitempty"`
  MyMetada    dispers.Metadata `json:"metadata,omitempty"` // A changer pour différencier chaque acteur
}

func (t *queryDoc) ID() string {
	return t.queryID
}

func (t *queryDoc) Rev() string {
	return t.queryRev
}

func (t *queryDoc) DocType() string {
	return "io.cozy.ml"
}

func (t *queryDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

func (t *queryDoc) SetID(id string) {
	t.queryID = id
}

func (t *queryDoc) SetRev(rev string) {
	t.queryRev = rev
}

func NewQueryDoc(queryID string, queryRev string, MyTraining Training, MyMetada dispers.Metadata) *queryDoc {
	return &queryDoc{
    queryID: queryID,
    queryRev: queryRev,
    MyTraining: MyTraining,
    MyMetada: MyMetada,
    }
}

func GetTrainingState(id string) echo.Map {
  couchdb.EnsureDBExist(prefixer.ConductorPrefixer, "io.cozy.ml")
  fetched := &queryDoc{}
  err := couchdb.GetDoc(prefixer.ConductorPrefixer, "io.cozy.ml", id, fetched)
  if err != nil {
    return echo.Map{"outcome": "error",
                  "message": err }
  }
  return echo.Map{"outcome": "ok",
                  "training" : fetched.MyTraining,
                  "metadata" : fetched.MyMetada}
}

/*
In order to handle several layers of DA, we create a structure called AggregationLayer
It is pretty much the same than layers in Neural Networks.
type aggregationLayer struct {
  input            string
  output           string
  unit             int16
  process          string
  dataaggregators  []actor
}
*/

type conductor struct {
  doc                 queryDoc // Doc in the querier's database where are saved parameters, metadata and results
  mPrefixer           prefixer.Prefixer
  targetfinders       []actor
  conceptindexors     []actor
  datas               []actor
  dataaggregators     []actor
  maindataaggregator  []actor
  MyTraining           Training
  /*stackAggr           []aggregationLayer*/
}

// NewConductor returns a Conductor object with the specified values.
// This object will be created directly in the cmd shell / web api
// This object use the major part of what have been created before in this script
func NewConductor(domain, prefix string) *conductor {

  pref := prefixer.NewPrefixer(domain, prefix)

  // Doc's creation in CouchDB
  couchdb.EnsureDBExist(prefixer.DataAggregatorPrefixer, "io.cozy.aggregation")

  /*
  doc := &DataAggrDoc{
    dataAggrDocID: "",
    dataAggrDocRev: "",
    Input: inputDA,
  }

 couchdb.CreateDoc(prefixer.DataAggregatorPrefixer, doc)
 */

  // récupérer l'id du doc sur prefix/io.cozy.ml
  doc_id := "17f78f7e8f7484z6"
  doc_rev := "2-46148"

	return &conductor{
    mPrefixer: pref,
    doc: queryDoc{
      queryID: doc_id,
      queryRev: doc_rev,
    },
  }
}

func (c *conductor) DecrypteConcept() dispers.Metadata { return nil }

func (c *conductor) ReachTargets() dispers.Metadata { return nil }

func (c *conductor) GetTrain() dispers.Metadata { return nil }

func (c *conductor) Aggregate() dispers.Metadata { return nil }

func (c *conductor) UpdateDoc(key string, metadata dispers.Metadata) error { return nil }

// This method is the most general. This is the only one used in CMD and Web's files. It will use the 5 previous methods to work
func (c *conductor) Lead() error {
  /*
  tempMetadata := dispers.NewMetadata("aujourd'hui", true)
  UpdateDoc("meta-task-0-init", tempMetadata)

  if (tempMetadata.Outcome()){
    tempMetadata = DecrypteConcept()
    UpdateDoc("meta-task-1-ci", tempMetadata)
  }

  if (tempMetadata.Outcome()){
    tempMetadata = ReachTargets()
    UpdateDoc("meta-task-2-tf", tempMetadata)
  }

  if (tempMetadata.Outcome()){
    tempMetadata = GetData()
    UpdateDoc("meta-task-3-d", tempMetadata)
  }

  if (tempMetadata.Outcome()){
    tempMetadata = Aggregate()
    UpdateDoc("meta-task-4-da", tempMetadata)
  }
*/

  return nil

}
