package couchdb

import (
	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

// ContactByEmail is used to find a contact by its email address
var ContactByEmail = &View{
	Name:    "contacts-by-email",
	Doctype: consts.Contacts,
	Map: `
function(doc) {
	if (isArray(doc.email)) {
		for (var i = 0; i < doc.email.length; i++) {
			emit(doc.email[i].address, doc._id);
		}
	}
}
`,
}

// Views is the list of all views that are created by the stack.
var Views = []*View{
	ContactByEmail,
}

// ViewsByDoctype returns the list of views for a specified doc type.
func ViewsByDoctype(doctype string) []*View {
	var views []*View
	for _, view := range Views {
		if view.Doctype == doctype {
			views = append(views, view)
		}
	}
	return views
}

// globalIndexes is the index list required on the global databases to run
// properly.
var globalIndexes = []*mango.Index{
	mango.IndexOnFields(consts.Exports, "by-domain", []string{"domain", "created_at"}),
}

// ConceptIndexorIndexes is the index list required by an instance to run properly.
var conceptIndexorIndexes = []*mango.Index{
	mango.IndexOnFields("io.cozy.hashconcept", "concept-index", []string{"concept"}),
}

// ConductorIndexes is the index list required by an instance to run properly.
var conductorIndexes = []*mango.Index{
	mango.IndexOnFields("io.cozy.instances", "hash", []string{"hash"}),
	mango.IndexOnFields("io.cozy.async", "async-task", []string{"queryid", "layerid", "daid"}),
	mango.IndexOnFields("io.cozy.async", "async-tasks", []string{"queryid", "layerid"}),
}

// secretIndexes is the index list required on the secret databases to run
// properly
var secretIndexes = []*mango.Index{
	mango.IndexOnFields(consts.AccountTypes, "by-slug", []string{"slug"}),
}

// DomainAndAliasesView defines a view to fetch instances by domain and domain
// aliases.
var DomainAndAliasesView = &View{
	Name:    "domain-and-aliases",
	Doctype: consts.Instances,
	Map: `
function(doc) {
  emit(doc.domain);
  if (isArray(doc.domain_aliases)) {
    for (var i = 0; i < doc.domain_aliases.length; i++) {
      emit(doc.domain_aliases[i]);
    }
  }
}
`,
}

// globalViews is the list of all views that are created by the stack on the
// global databases.
var globalViews = []*View{
	DomainAndAliasesView,
}

// InitGlobalDB defines views and indexes on the global databases. It is called
// on every startup of the stack.
func InitGlobalDB() error {
	if err := DefineIndexes(GlobalSecretsDB, secretIndexes); err != nil {
		return err
	}
	if err := DefineIndexes(GlobalDB, globalIndexes); err != nil {
		return err
	}
	if err := DefineIndexes(prefixer.ConceptIndexorPrefixer, conceptIndexorIndexes); err != nil {
		return err
	}
	if err := DefineIndexes(prefixer.ConductorPrefixer, conductorIndexes); err != nil {
		return err
	}
	if err := DefineIndexes(prefixer.TestConceptIndexorPrefixer, conceptIndexorIndexes); err != nil {
		return err
	}
	if err := DefineIndexes(prefixer.TestConductorPrefixer, conductorIndexes); err != nil {
		return err
	}
	return DefineViews(GlobalDB, globalViews)
}
