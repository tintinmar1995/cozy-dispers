package aggregations

import (
	"math"
	"math/rand"
	"regexp"
	"strings"

	"github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/james-bowman/nlp"
	"gonum.org/v1/gonum/mat"
)

func getAmountTag(amount float64) string {
	if amount < -550 {
		return "tag_v_b_expense"
	} else if amount < -100 {
		return "tag_b_expense"
	} else if amount < -20 {
		return "tag_expense"
	} else if amount < 0 {
		return "tag_noise_neg"
	} else if amount < 50 {
		return "tag_noise_pos"
	} else if amount < 200 {
		return "tag_income"
	} else if amount < 1200 {
		return "tag_b_income"
	}
	return "tag_activity_income"
}

func getSignTag(amount float64) string {
	if amount < 0 {
		return "tag_neg"
	} else {
		return "tag_pos"
	}
}

func getSanitizedLabel(label string) string {

	label = strings.ToLower(label)

	// TODO: Deal with non-ASCII chars

	// replace date by tag_date
	dateRegex := regexp.MustCompile(`(\d{1,2}[/]\d{1,2}[/]\d{2,4})|(\d{1,2}[/]\d{1,2})|(\d{1,2}[-]\d{1,2})`)
	label = dateRegex.ReplaceAllString(label, " tag_date ")

	// build month tagging lambda function
	for _, word := range []string{"janvier", "janv", "fevrier", "fev", "mars", "mar", "avril", "avr", "mai", "juin", "juillet", "juil", "aout", "aou", "septembre", "sept", "octobre", "oct", "novembre", "nov", "decembre", "dec"} {
		label = strings.ReplaceAll(label, word, "tag_month")
	}

	// delete unnecessary chars
	unnecessaryChars := regexp.MustCompile(`[^a-zA-Z_ ]`)
	label = unnecessaryChars.ReplaceAllString(label, "")

	// replace double spaces with simple spaces
	label = strings.ReplaceAll(label, "  ", " ")

	return label
}

// Preprocessing apply a preprocessing over data before training on it
// Preprocessing takes as argument the doctype to choose which preprocess to use
// TODO : Broke the preprocess function in pieces to create more atomic functions
func Preprocessing(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	if err := needArgs(args, "voc", "doctype"); err != nil {
		return nil, err
	}
	if _, ok := result["preprocessed_data"]; !ok {
		result["preprocessed_data"] = []map[string]interface{}{}
	}

	// TODO : Preprocessing is applying one preprocess per doctype, this is a bad way
	if args["doctype"].(string) == "io.cozy.bank.operations" {

		res := make(map[string]interface{})

		// Converting each word of vocabulary as a unary-gram token
		// TODO: Fork James Bowman's repo to add n-grams
		vec := nlp.NewCountVectoriser()
		vec.Fit(args["voc"].(string))

		// Cleaning up row's label
		label := getSanitizedLabel(row["label"].(string))
		label = label + " " + getSignTag(row["amount"].(float64))
		label = label + " " + getAmountTag(row["amount"].(float64))

		features, err := vec.Transform(label)
		if err != nil {
			return nil, err
		}
		res["features"] = features

		// Setting the boolean value to predict
		res["truth"] = row[args["target_key"].(string)] == args["target_value"].(string)
		result["preprocessed_data"] = append(result["preprocessed_data"].([]map[string]interface{}), res)

	}

	return result, nil
}

// hypothesisFunction returns 1 /(1 + exp(-theta*features))
func hypothesisFunction(theta []float64, features mat.Matrix) float64 {
	out := 0.0
	for index := range theta {
		out = out - theta[index]*features.At(index, 0)
	}
	out = 1 + math.Exp(out)
	return 1 / out
}

// LogisticRegressionMap returns Gradient (Resp. & Hessian) to compute SGD (Resp. Newton-Raphson)
// Inspired from https://papers.nips.cc/paper/3150-map-reduce-for-machine-learning-on-multicore.pdf
// and https://www.internalpointers.com/post/cost-function-logistic-regression
// In both case, gradient should be computed on the data received by the mapper
func LogisticRegressionMap(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	if err := needArgs(args, "optimize"); err != nil {
		return nil, err
	}

	features := row["features"].(mat.Matrix)
	lenFeatures, _ := features.Dims()

	// Check if the parameters of the Logistic Regression are set as args
	// If not, parameters are initialized at random
	var theta []float64
	if _, ok := args["theta"]; !ok {
		theta = make([]float64, lenFeatures)
		random := rand.New(rand.NewSource(0))
		for index := range theta {
			theta[index] = random.Float64()
		}
		result["theta"] = theta
	} else {
		theta = args["theta"].([]float64)
		if len(theta) != lenFeatures {
			return nil, errors.ErrLengthConsistency
		}
	}

	// Initialize results if needed
	if _, ok := result["gradient"]; !ok {
		result["gradient"] = make([]float64, lenFeatures)
	}
	if _, ok := result["hessian"]; !ok {
		result["hessian"] = make(map[[2]int]float64)
	}

	// Convert truth as float64 and get prediction
	truth, err := asFloat64(row["truth"])
	if err != nil {
		return nil, err
	}
	prediction := hypothesisFunction(theta, features)

	// Compute gradient and hessian (if Newton Raphson method is chosen)
	i := 0
	for i < lenFeatures {
		j := i // hessian is symmetric
		result["gradient"].([]float64)[i] = result["gradient"].([]float64)[i] + (truth-prediction)*features.At(i, 0)
		for args["optimize"].(string) == "nr" && j < lenFeatures {
			value := result["hessian"].(map[[2]int]float64)[[2]int{i, j}] + prediction*(prediction-1)*features.At(i, 0)*features.At(j, 0)
			result["hessian"].(map[[2]int]float64)[[2]int{i, j}] = value
			j = j + 1
		}
		i = i + 1
	}

	return result, nil
}

// LogisticRegressionReduce sums up Gradient and Hessian (if Newton Raphson)
func LogisticRegressionReduce(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// Initialize results if needed
	if _, ok := result["gradient"]; !ok {
		result["gradient"] = make([]float64, len(row["gradient"].([]float64)))
	}
	if _, ok := result["hessian"]; !ok {
		result["hessian"] = make(map[[2]int]float64)
	}

	// sum up gradient and hessian
	i := 0
	for i < len(row["gradient"].([]float64)) {
		j := i
		result["gradient"].([]float64)[i] = result["gradient"].([]float64)[i] + row["gradient"].([]float64)[i]
		for args["optimize"].(string) == "nr" && j < len(row["gradient"].([]float64)) {
			value := result["hessian"].(map[[2]int]float64)[[2]int{i, j}] + row["hessian"].(map[[2]int]float64)[[2]int{i, j}]
			result["hessian"].(map[[2]int]float64)[[2]int{i, j}] = value
			j = j + 1
		}
		i = i + 1
	}

	return result, nil
}

// LogisticRegressionUpdateParameters updates the Logistic Regression parameters
// The function is capable of updating with two methods :
//       - Gradient Descent method
//       - Newton-Raphson method
func LogisticRegressionUpdateParameters(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// Check if args are present and Initialize variables
	if err := needArgs(args, "theta", "optimize"); err != nil {
		return nil, err
	}
	lengthFeatures := len(row["gradient"].([]float64))
	gradient := make([]float64, lengthFeatures)
	hessian := mat.NewSymDense(lengthFeatures, nil)
	theta := args["theta"].([]float64)
	newTheta := make([]float64, len(theta))

	// Devide each coef of gradient and hessian (if Newton Raphson) by length
	i := 0
	for i < lengthFeatures {
		j := i
		gradient[i] = row["gradient"].([]float64)[i] / float64(row["length"].(int))
		for args["optimize"].(string) == "nr" && j < lengthFeatures {
			value := row["hessian"].(map[[2]int]float64)[[2]int{i, j}] / float64(row["length"].(int))
			hessian.SetSym(i, j, value)
			j = j + 1
		}
		i = i + 1
	}

	if args["optimize"].(string) == "nr" {

		// theta := theta − H−1 * gradient
		invHessian := mat.NewSymDense(lengthFeatures, nil)
		err := invHessian.PowPSD(hessian, -1)
		if err != nil {
			return nil, err
		}
		// TODO : compute invHessian * gradient
		for index := range theta {
			newTheta[index] = theta[index] - gradient[index]
		}
		result["theta"] = newTheta
	} else if args["optimize"].(string) == "gd" {

		// theta := theta − gradient
		for index := range theta {
			newTheta[index] = theta[index] - gradient[index]
		}
		result["theta"] = newTheta
	}

	return result, nil
}
