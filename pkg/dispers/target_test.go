package enclave

import (
	"testing"
	"time"

	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {

	inst := dispers.Instance{
		Domain: "prettyname4acozy.mycozy.cloud",
		Token:  dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
	}

	localQuery := dispers.LocalQuery{
		FindRequest: map[string]interface{}{},
	}

	in := dispers.InputT{
		LocalQuery: localQuery,
	}
	out := buildQuery(inst, in.LocalQuery)
	assert.Equal(t, inst.Domain, out.Domain)
}

func TestCleanTargetsList(t *testing.T) {
	insts := []dispers.Instance{
		dispers.Instance{
			Domain:           "prettyname4acozy.mycozy.cloud",
			SubscriptionDate: time.Now(),
			Token:            dispers.Token{TokenBearer: "itsagoodone"},
		}, dispers.Instance{
			Domain:           "",
			SubscriptionDate: time.Now(),
			Token:            dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
		}, dispers.Instance{
			SubscriptionDate: time.Now(),
			Token:            dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
		}, dispers.Instance{
			SubscriptionDate: time.Now(),
			Domain:           "notaken.mycozy.cloud",
		}, dispers.Instance{
			Domain:           "emptytoken.mycozy.cloud",
			SubscriptionDate: time.Now(),
			Token:            dispers.Token{},
		}, dispers.Instance{
			Domain:           "anothergoodcozy.mycozy.cloud",
			SubscriptionDate: time.Now(),
			Token:            dispers.Token{TokenBearer: "itsagoodone"},
		}, dispers.Instance{
			Domain: "nosubscriptiondate.mycozy.cloud",
			Token:  dispers.Token{TokenBearer: "vdf5s4fs2ffse4fc7es5fd"},
		}, dispers.Instance{
			Domain:           "yesterdayisagoodsongfromthebeatles.mycozy.cloud",
			Token:            dispers.Token{TokenBearer: "thebadone"},
			SubscriptionDate: time.Now().AddDate(0, 0, -1),
		}, dispers.Instance{
			Domain:           "yesterdayisagoodsongfromthebeatles.mycozy.cloud",
			Token:            dispers.Token{TokenBearer: "itsagoodone"},
			SubscriptionDate: time.Now(),
		}, dispers.Instance{
			Domain:           "back2thefuture.mycozy.cloud",
			Token:            dispers.Token{TokenBearer: "itsagoodone"},
			SubscriptionDate: time.Now().AddDate(0, 0, 2),
		}, dispers.Instance{
			Domain:           "back2thefuture.mycozy.cloud",
			Token:            dispers.Token{TokenBearer: "thebadone"},
			SubscriptionDate: time.Now(),
		},
	}
	cleanTargetsList(&insts)
	assert.Equal(t, 4, len(insts))
	for _, inst := range insts {
		assert.Equal(t, "itsagoodone", inst.Token.TokenBearer)
	}
}

func TestCleanEmptyTargetsList(t *testing.T) {
	insts := []dispers.Instance{}
	cleanTargetsList(&insts)
	assert.Equal(t, 0, len(insts))
}
