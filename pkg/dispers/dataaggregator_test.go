package enclave

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/stretchr/testify/assert"
)

func TestCategPreprocessing(t *testing.T) {

	// Get data from dummy_dataset
	s := ""
	absPath, _ := filepath.Abs("../../assets/test/dummy_bank_data.json")
	buf, err := ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	s = string(buf)
	assert.Equal(t, "[\n    {\n      \"", s[:15])
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(s), &data)
	assert.NoError(t, err)

	voc := ""
	absPath, _ = filepath.Abs("../../assets/test/vocabulary.txt")
	buf, err = ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	voc = string(buf)

	args := make(map[string]interface{})
	args["voc"] = voc
	args["target_key"] = "cozyCategoryId"
	args["target_value"] = "400340"
	args["doctype"] = "io.cozy.bank.operations"
	encData, _ := json.Marshal(data)
	encJob, _ := json.Marshal([]query.AggregationJob{query.AggregationJob{
		Job:  "preprocess",
		Args: args,
	}})
	in := query.InputDA{
		EncryptedData: encData,
		EncryptedJobs: encJob,
	}
	_, err = AggregateData(in)
	assert.NoError(t, err)
}

func TestAggregateOneLayer(t *testing.T) {
	// Get Data From dummy_dataset
	s := ""
	absPath, _ := filepath.Abs("../../assets/test/dummy_dataset.json")
	buf, err := ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	s = string(buf)
	assert.Equal(t, s[:29], "[\n {\n   \"sepal_length\": 5.1,\n")
	var data []map[string]interface{}
	json.Unmarshal([]byte(s), &data)

	// Test Sum
	results := make(map[string]interface{})
	args := make(map[string]interface{})
	function := "sum"
	for idx, rowData := range data {
		args["key"] = "sepal_length"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
		args["key"] = "sepal_width"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
	}
	assert.Equal(t, map[string]interface{}{"sum_sepal_length": 876.5000000000002, "sum_sepal_width": 458.10000000000014}, results)

	// Test SumSquare
	results = make(map[string]interface{})
	args = make(map[string]interface{})
	args["keys"] = []string{"sepal_length", "sepal_width"}
	function = "sum_square"
	for idx, rowData := range data {
		args["key"] = "sepal_length"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
		args["key"] = "sepal_width"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
	}
	assert.Equal(t, map[string]interface{}{"sum_square_sepal_length": 5223.849999999998, "sum_square_sepal_width": 1427.049999999999}, results)

	// Test Min
	results = make(map[string]interface{})
	args = make(map[string]interface{})
	function = "min"
	for idx, rowData := range data {
		args["key"] = "sepal_length"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
		args["key"] = "sepal_width"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
	}
	assert.Equal(t, map[string]interface{}{"min_sepal_length": 4.3, "min_sepal_width": 2.0}, results)

	// Test Max
	results = make(map[string]interface{})
	args = make(map[string]interface{})
	function = "max"
	for idx, rowData := range data {
		args["key"] = "sepal_length"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
		args["key"] = "sepal_width"
		err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
		assert.NoError(t, err)
	}
	assert.Equal(t, map[string]interface{}{"max_sepal_length": 7.9, "max_sepal_width": 4.4}, results)

}

func TestAggregateMean(t *testing.T) {
	// Get Data From dummy_dataset
	res := make([]map[string]interface{}, 4)

	i := 0
	for i < 4 {
		absPath, _ := filepath.Abs(strings.Join([]string{"../../assets/test/dummy_dataset", strconv.Itoa(i), ".json"}, ""))
		buf, _ := ioutil.ReadFile(absPath)
		s := string(buf)

		var data []map[string]interface{}
		json.Unmarshal([]byte(s), &data)

		results := make(map[string]interface{})
		args := make(map[string]interface{})
		function := "sum"
		for idx, rowData := range data {
			args["key"] = "sepal_length"
			err := applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
			assert.NoError(t, err)
			args["key"] = "sepal_width"
			err = applyAggregateFunction(idx, &results, rowData, query.AggregationFunction{Function: function, Args: args})
			assert.NoError(t, err)
		}
		results["length"] = len(data)
		res[i] = results
		i = i + 1
	}
	assert.Equal(t, []map[string]interface{}{map[string]interface{}{"length": 7, "sum_sepal_length": 34.3, "sum_sepal_width": 23.699999999999996}, map[string]interface{}{"length": 21, "sum_sepal_length": 106.6, "sum_sepal_width": 73.19999999999999}, map[string]interface{}{"length": 37, "sum_sepal_length": 198.99999999999997, "sum_sepal_width": 115.7}, map[string]interface{}{"length": 85, "sum_sepal_length": 536.5999999999998, "sum_sepal_width": 245.50000000000003}}, res)

	encData, _ := json.Marshal(res)
	aggrJobs := []query.AggregationJob{
		query.AggregationJob{
			Job: "mean",
			Args: map[string]interface{}{
				"sum": "sum_sepal_width",
			},
		},
		query.AggregationJob{
			Job: "mean",
			Args: map[string]interface{}{
				"sum": "sum_sepal_length",
			},
		},
	}
	encJob, _ := json.Marshal(aggrJobs)
	in2 := query.InputDA{
		EncryptedData: encData,
		EncryptedJobs: encJob,
	}
	means, err := AggregateData(in2)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"length": 4, "sum_sum_sepal_length": 21.667510031039438, "sum_sum_sepal_width": 12.886690892573245}, means)
}
