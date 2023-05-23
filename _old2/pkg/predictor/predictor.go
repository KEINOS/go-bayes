package predictor

import (
	"fmt"

	"github.com/KEINOS/go-bayes/pkg/dataset"
	"github.com/KEINOS/go-bayes/pkg/theorem"
	"github.com/KEINOS/go-bayes/pkg/trainer"
)

type Predictor struct {
	TrainedDB trainer.Trainer
	ClassDB   dataset.DB
}

type Predicted struct {
	ClassID  uint64
	Prob     float64
	RawValue any
}

func New(trainedDB trainer.Trainer, classDB dataset.DB) *Predictor {
	return &Predictor{
		TrainedDB: trainedDB,
		ClassDB:   classDB,
	}
}

func (p *Predictor) Predict(fromA uint64) ([]Predicted, error) {
	predicted := []Predicted{}

	for p.ClassDB.HasNext() {
		toClassID, anyValue, err := p.ClassDB.Next()
		if err != nil {
			return nil, err
		}

		prob := p.predict(fromA, toClassID)

		predicted = append(predicted, Predicted{
			ClassID:  toClassID,
			Prob:     prob,
			RawValue: anyValue,
		})
	}

	return predicted, nil
}

func (p *Predictor) predict(fromA uint64, toB uint64) float64 {
	priorPtoB := p.TrainedDB.PriorPtoB(toB)
	priorPfromAtoB := p.TrainedDB.PriorPfromAtoB(fromA, toB)
	priorPNotFromAtoB := p.TrainedDB.PriorPNotFromAtoB(fromA, toB)

	fmt.Println("FromA:", fromA, "ToB:", toB)
	fmt.Println("  PriorPtoB:", priorPtoB)
	fmt.Println("  PriorPfromAtoB:", priorPfromAtoB)
	fmt.Println("  PriorPNotFromAtoB:", priorPNotFromAtoB)

	return theorem.Bayes(priorPtoB, priorPfromAtoB, priorPNotFromAtoB)
}
