# Iris Example

This runnable program demonstrates the core `go-bayes` flow:

1. Create a predictor.
2. Train it with ordered transitions.
3. Predict the next class ID from an ordered context.
4. Resolve that ID to the value stored during training.

Each training sequence has this form:

```text
sepal length -> sepal width -> petal length -> petal width -> species
```

## Run

From the repository root:

```sh
go run ./_examples/iris
```

Expected output:

```text
trained: 150 samples
5.1, 3.5, 1.4, 0.2 -> Iris-setosa
7.0, 3.2, 4.7, 1.4 -> Iris-versicolor
6.3, 3.3, 6.0, 2.5 -> Iris-virginica
```

The example keeps each decimal measurement as its canonical CSV string so its fractional part remains part of the item ID. The queried rows are present in the training data. For each query, the expected stored value is a species name. The current predictor folds the exact ordered measurement context into an ID; it does not calculate distance between measurements or classify an unseen flower by similarity. This example therefore demonstrates context-to-class transition prediction, not a holdout accuracy benchmark.

## Dataset

[`iris.data`](iris.data) contains the original comma-separated records from the official UCI archive. It does not use the JSON conversion under the repository's `testdata/` directory.

- Dataset page: [Iris, UCI Machine Learning Repository](
  https://archive.ics.uci.edu/dataset/53/iris)
- Official archive: [iris.zip](
  https://archive.ics.uci.edu/static/public/53/iris.zip)
- DOI: [10.24432/C56C76](https://doi.org/10.24432/C56C76)
- License: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)

Suggested citation: Fisher, R. A. (1936). *Iris* [Dataset]. UCI Machine Learning Repository. DOI: 10.24432/C56C76.
