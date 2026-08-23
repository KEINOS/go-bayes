# Mushroom Example

This program uses 22 categorical features to recover the stored edible or
poisonous class. It shows that predictor inputs do not need to be numbers.

Each training sequence has this form:

```text
22 categorical features -> edible or poisonous
```

The source file uses one-character codes. For example, `x` means a convex cap,
`p` means a pungent odor, and `u` means an urban habitat. The `?` value means
that the stalk root is missing. The predictor handles it as another input
token.

## Run

From the repository root:

```sh
go run ./_examples/mushroom
```

Expected output:

```text
trained: 8124 samples
cap=x, odor=p, stalk-root=e, habitat=u -> poisonous
cap=x, odor=a, stalk-root=c, habitat=g -> edible
cap=x, odor=n, stalk-root=?, habitat=w -> edible
```

The queried records are present in the training data. The predictor matches
their exact ordered contexts. This is an API example, not a holdout accuracy
test or a similarity-based classifier.

> [!WARNING]
> Never use this example to decide whether a real mushroom is safe to eat. The
> dataset documentation states that there is no simple rule for this decision.

## Dataset

[`agaricus-lepiota.data`](agaricus-lepiota.data) is the original
comma-separated file from the UCI archive.

- Dataset page: [Mushroom, UCI Machine Learning Repository](
  https://archive.ics.uci.edu/dataset/73/mushroom)
- Official archive: [mushroom.zip](
  https://archive.ics.uci.edu/static/public/73/mushroom.zip)
- DOI: [10.24432/C5959T](https://doi.org/10.24432/C5959T)
- License: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)

Suggested citation: *Mushroom* [Dataset]. (1987). UCI Machine Learning
Repository. DOI: 10.24432/C5959T.
