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

func TestAggregateSum(t *testing.T) {
	// Get Data From dummy_dataset
	s := ""
	absPath, _ := filepath.Abs("../../assets/test/dummy_dataset.json")
	buf, err := ioutil.ReadFile(absPath)
	assert.NoError(t, err)
	s = string(buf)
	assert.Equal(t, s[:29], "[\n {\n   \"sepal_length\": 5.1,\n")
	var data []map[string]interface{}
	json.Unmarshal([]byte(s), &data)

	args := make(map[string]interface{})
	args["keys"] = []string{"sepal_length", "sepal_width"}
	function := "sum"

	res, err := applyAggregateFunction(data, query.AggregationFunction{Function: function, Args: args})
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"length": 150, "sepal_length": 876.5000000000002, "sepal_width": 458.10000000000014}, res)

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

		args := make(map[string]interface{})
		args["keys"] = []string{"sepal_length", "sepal_width"}
		function := "sum"

		results, _ := applyAggregateFunction(data, query.AggregationFunction{Function: function, Args: args})
		res[i] = results
		i = i + 1
	}

	assert.Equal(t, []map[string]interface{}([]map[string]interface{}{map[string]interface{}{"length": 7, "sepal_length": 34.3, "sepal_width": 23.699999999999996}, map[string]interface{}{"length": 21, "sepal_length": 106.6, "sepal_width": 73.19999999999999}, map[string]interface{}{"length": 37, "sepal_length": 198.99999999999997, "sepal_width": 115.7}, map[string]interface{}{"length": 85, "sepal_length": 536.5999999999998, "sepal_width": 245.50000000000003}}), res)

	args := make(map[string]interface{})
	args["keys"] = []string{"sepal_length", "sepal_width"}
	args["weight"] = "length"
	encData, _ := json.Marshal(res)
	encFunc, _ := json.Marshal(query.AggregationFunction{
		Function: "sum",
		Args:     args,
	})
	in2 := query.InputDA{
		EncryptedData:     encData,
		EncryptedFunction: encFunc,
	}
	means, err := AggregateData(in2)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}(map[string]interface{}{"length": 4, "sepal_length": 21.667510031039438, "sepal_width": 12.886690892573245}), means)

}
