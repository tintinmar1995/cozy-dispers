package aggregations

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocessing(t *testing.T) {

	// Get data from dummy_dataset
	s := ""
	absPath, _ := filepath.Abs("../../../assets/test/dummy_bank_data.json")
	buf, err := ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	s = string(buf)
	assert.Equal(t, "[\n    {\n      \"", s[:15])
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(s), &data)
	assert.NoError(t, err)

	voc := ""
	absPath, _ = filepath.Abs("../../../assets/test/vocabulary.txt")
	buf, err = ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	voc = string(buf)

	results := make(map[string]interface{})
	args := make(map[string]interface{})
	args["voc"] = voc
	args["target_key"] = "cozyCategoryId"
	args["target_value"] = "400340"
	args["optimize"] = "gd"
	args["doctype"] = "io.cozy.bank.operations"
	for _, rowData := range data {
		results, err = Preprocessing(results, rowData, args)
		assert.NoError(t, err)
	}

	results2 := make(map[string]interface{})
	for _, rowData := range results["preprocessed_data"].([]map[string]interface{})[0:15] {
		results2, err = LogisticRegressionMap(results2, rowData, args)
		assert.NoError(t, err)
	}

	args["theta"] = results2["theta"].([]float64)

	results3 := make(map[string]interface{})
	for _, rowData := range results["preprocessed_data"].([]map[string]interface{})[15:40] {
		results3, err = LogisticRegressionMap(results3, rowData, args)
		assert.NoError(t, err)
	}

	results4 := make(map[string]interface{})
	for _, rowData := range results["preprocessed_data"].([]map[string]interface{})[40:] {
		results4, err = LogisticRegressionMap(results4, rowData, args)
		assert.NoError(t, err)
	}

	results5 := make(map[string]interface{})
	for _, rowData := range []map[string]interface{}{results2, results3, results4} {
		results5, err = LogisticRegressionReduce(results5, rowData, args)
		assert.NoError(t, err)
	}
	results5["length"] = 3

	results6 := make(map[string]interface{})
	_, err = LogisticRegressionUpdateParameters(results6, results5, args)
	assert.NoError(t, err)

}
