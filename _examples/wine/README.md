# Wine Example

This program uses chemical measurements to recover the stored cultivar class.
It demonstrates a longer ordered context than the Iris example.

Each training sequence has this form:

```text
13 chemical measurements -> cultivar class
```

All 13 measurements are used for prediction. The output shows only alcohol,
color intensity, and proline to keep each line short.

## Run

From the repository root:

```sh
go run ./_examples/wine
```

Expected output:

```text
trained: 178 samples
alcohol=14.23, color=5.64, proline=1065 -> cultivar 1
alcohol=12.37, color=1.95, proline=520 -> cultivar 2
alcohol=12.86, color=4.1, proline=630 -> cultivar 3
```

The queried records are present in the training data. The predictor matches
their exact ordered contexts. This is an API example, not a holdout accuracy
test or a similarity-based wine classifier.

## Dataset

[`wine.data`](wine.data) is the original comma-separated file from the UCI
archive.

- Dataset page: [Wine, UCI Machine Learning Repository](
  https://archive.ics.uci.edu/dataset/109/wine)
- Official archive: [wine.zip](
  https://archive.ics.uci.edu/static/public/109/wine.zip)
- DOI: [10.24432/C5PC7J](https://doi.org/10.24432/C5PC7J)
- License: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)

Suggested citation: Aeberhard, S., & Forina, M. (1992). *Wine* [Dataset]. UCI
Machine Learning Repository. DOI: 10.24432/C5PC7J.
