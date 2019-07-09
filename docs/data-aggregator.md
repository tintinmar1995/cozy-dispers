# Cozy-DISPERS : Data Aggregator

The main goal of the whole Cozy-DISPERS process is to conduct data to Data Aggregator with privacy-by-design. When Data Aggregator receives data, this goal is accomplished.

Data Aggregators are made to compute the query from the data. Since a large set of data can bring pieces of information about entities, Cozy-DISPERS devides the computation process in several process made by several Data Aggregators. By doing so, we parallelize and fasten the process.


| Available Aggr. Func.    |  Index   |   Args[0]  | Args[1]    |
| ------------- |: ------: | ---------: | ---------: |
| Weighted-Sum  |   `sum`    | Targeted keys | Weight's key  |


## How-to compute a weighted-sum

**Step 1** : Choose a dataset to compute the query on

In our case, the dataset is a part of the [iris database](https://www.iris-database.org). Each row comes from a fake DocType io.cozy.iris in an instance of Cozy.

```json
{
   "sepal_length": 5.1,
   "sepal_width": 3.5,
   "petal_length": 1.4,
   "petal_width": 0.2,
   "species": "setosa"
 },
 {
   "sepal_length": 4.9,
   "sepal_width": 3,
   "petal_length": 1.4,
   "petal_width": 0.2,
   "species": "setosa"
 }
```

**Step 2** : Compute sums in subsets

Let's split our set in four parts, each part will be given to one instance of Data Aggregator with this InputDA struct.

In this case, we will compute a sum weighted by the petal_width.

```golang
args["keys"] := []string{"sepal_length", "sepal_width"}
args["weight"] = "petal_width"

in := dispers.InputDA{
		Data: data1,
		Job: dispers.AggregationFunction{
			Function: "sum",
			Args:     args,
		},
}
```


**Step 3** : Make the request & Get the result

```http
POST query/dataaggregator/aggregate HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputDA

> []float64{876.5000000000002, 458.10000000000014}

```


## Aggregation with Several Layers

Data Aggregator (DA) has been desgined to be used sequantially. You can organize n DAs in layers.

### Examples :

#### Weighted-mean

| Layer  |  Size      |   Role                    |
| ------ |: --------: | ------------------------: |
| 1      |   5        | Compute sums on 5 subsets |
| 2      |   1        | Compute mean from 5 sums  |
> Return : mean for each variable

#### Naive Bayes

| Layer  |  Size       |   Role                    |
| ------ |: ---------: | ------------------------: |
| 1      |   10        | Preprocess data           |
| 2      |   15        | Train a Naive Bayes model |
| 3      |   1         | Merge Naive Bayes models  |
| 4      |   5         | Test Naive Bayes models   |
| 5      |   1         | Merge tests               |
> Return : Parameters and metrics (Accuracy, ...)
