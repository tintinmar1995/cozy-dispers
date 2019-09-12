package patches

// LogisticRegressionUpdateParameters updates the Logistic Regression parameters
import (
	"github.com/cozy/cozy-stack/pkg/dispers/aggregation"
	"gonum.org/v1/gonum/mat"
)

// The function is capable of updating with two methods :
//       - Gradient Descent method
//       - Newton-Raphson method
func LogisticRegressionUpdateParameters(result *map[string]interface{}, row map[string]interface{}, args map[string]interface{}) error {

	// Check if args are present and Initialize variables
	if err := aggregations.NeedArgs(args, "theta", "optimize"); err != nil {
		return err
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
			return err
		}
		// TODO : compute invHessian * gradient
		for index := range theta {
			newTheta[index] = theta[index] - gradient[index]
		}
		(*result)["theta"] = newTheta
	} else if args["optimize"].(string) == "gd" {

		// theta := theta − gradient
		for index := range theta {
			newTheta[index] = theta[index] - gradient[index]
		}
		(*result)["theta"] = newTheta
	}

	return nil
}
