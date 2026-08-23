# Technical Specification

## Purpose

`go-bayes` is a Bayesian inference package for ordered data. Its current model is a Folded Context Transition Predictor (FCTP).

The main idea is small: every learned fact becomes a pair of two fixed-width IDs.

```text
INPUT ID -> CLASS ID
```

`INPUT ID` can represent one token or an ordered context of many tokens. `CLASS ID` represents a possible next value. The storage layer therefore does not need a different data structure for a longer context.

In this document, “two-ID pair” means two values in one record. It does not mean binary data with only `0` and `1`.

## Core Data Path

The predictor converts input data in three steps:

```mermaid
flowchart LR
    A[ordered tokens] --> B[item IDs]
    B --> C[fold into one context ID]
    D[expected token] --> E[class ID]
    C --> F[INPUT ID → CLASS ID]
    E --> F
```

1. Convert each supported token to a `uint64` item ID.
2. Fold one or more ordered item IDs into one `uint64` context ID.
3. Store the context ID and expected class ID as one two-ID pair.

This conversion removes the original context length from the storage boundary. A one-token input and a one-thousand-token input can both be represented by one `uint64` input ID.

## Token IDs

The current implementation accepts common Go values such as strings, booleans, integers, and floating-point numbers.

- A string receives an item ID from the first 64 bits of its BLAKE3 hash.
- A boolean becomes `0` or `1`.
- An integer becomes a `uint64` while preserving its bit pattern.
- A floating-point number is converted to `uint64`. Its fractional part is lost.

These rules create item IDs. They are separate from the configurable context hasher.

## Context Folding

The predictor passes the ordered item IDs to a context hasher. The hasher returns one `uint64` context ID.

```text
[ID(A), ID(B), ID(C)] -> FOLD -> INPUT ID
```

xxHash3 is the default context hasher. BLAKE3 is also available. A custom hasher can implement the public `Hasher` interface.

The order is part of the input. `FOLD(A, B, C)` and `FOLD(C, A, B)` are different contexts in normal use.

The result is deterministic for the same input and hasher. It is not unique in the mathematical sense. Any fixed-width hash can have collisions.

## Learning

`Train` reads an ordered sequence from left to right. Each value after the first value becomes a possible class.

For this sequence:

```text
A -> B -> C -> D
```

the predictor records the direct previous-token transition and the folded suffix contexts. The records for class `D` are:

```text
ID(C)          -> ID(D)
FOLD(C)        -> ID(D)
FOLD(B, C)     -> ID(D)
FOLD(A, B, C)  -> ID(D)
```

It also creates records that predict `B` and `C`. Every record still has only two values: one input ID and one class ID.

This suffix expansion lets the same class collect evidence from contexts of different lengths. It also causes the number of stored records to grow as training sequences become longer. A future pruning feature can remove low-value records.

## Prediction

`Predict` converts all supplied tokens to item IDs and folds the complete ordered input into one context ID.

For input `A, B, C`, it uses:

```text
FOLD(A, B, C) -> INPUT ID
```

The predictor scores that input ID against every known class ID. It returns the class ID with the highest positive score. `GetClass` maps the returned ID to the original value saved during training.

The current predictor uses exact context IDs. It does not measure similarity between token values. It also does not retry shorter contexts when the complete context has no useful match. A caller can add this backoff by calling `Predict` again with shorter suffixes.

If no class has a positive score, `Predict` returns `0`. Because `0` can also be a valid class ID, callers cannot use the returned ID alone to distinguish these cases. If multiple classes have the same highest score, any one of them can be returned.

## Bayesian Scoring

The in-memory logger counts observed `INPUT ID -> CLASS ID` pairs. During prediction, it derives frequencies from those counts and sends them to the current Bayes-based scoring helper. The predictor compares the returned scores; it does not expose a complex probability model to the storage interface.

This makes Bayesian scoring one part of the implementation. FCTP describes how the input context becomes an ID and how that ID connects to a class. The model is not Naive Bayes and does not split the input into independent features.

The probability input contract is under review before `v1.0.0`. Some current inputs are frequencies over all stored observations, rather than conditional probabilities. Any correction must define the intended behavior and add tests before it changes prediction results.

## Class Recovery

A hash is not reversible. The predictor therefore keeps a class map:

```text
CLASS ID -> original Go value
```

`GetClass` reads this map. It returns `nil` for an unknown ID.

JSON persistence stores the class map, but JSON restores numeric values as `float64`. JSON also does not store the selected context hasher. A restored predictor must use the same context hasher that was used for training.

## Design Boundary

The stable idea of FCTP is the two-ID learning boundary:

```text
arbitrary ordered context -> fixed-width INPUT ID
possible next value       -> fixed-width CLASS ID
learned fact               -> INPUT ID, CLASS ID
```

Hash algorithms, probability scoring, storage, pruning, and backoff can change without removing this idea. An implementation that no longer folds an ordered context into one input ID uses a different model.

The project is still before `v1.0.0`. API clarity, ease of use, and measured performance are more important than backward compatibility during this stage.
