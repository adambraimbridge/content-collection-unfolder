package differ

import "github.com/Workiva/go-datastructures/set"

type CollectionsDiffer interface {
	SymmetricDifference(incomingCollectionUuids []string, oldCollectionUuids []string) *set.Set
}

type defaultCollectionsDiffer struct {
}

func NewDefaultCollectionsDiffer() *defaultCollectionsDiffer {
	return &defaultCollectionsDiffer{}
}

func (dcd *defaultCollectionsDiffer) SymmetricDifference(incomingCollectionUuids []string, oldCollectionUuids []string) *set.Set {
	symDiffSet := set.New()

	relativeComplement(incomingCollectionUuids, oldCollectionUuids, symDiffSet)
	relativeComplement(oldCollectionUuids, incomingCollectionUuids, symDiffSet)

	return symDiffSet
}

func relativeComplement(firstCollection []string, secondCollection []string, aggregatingSet *set.Set) {
	secondColSet := set.New()
	for _, secondColUuid := range secondCollection {
		secondColSet.Add(secondColUuid)
	}

	for _, firstColUuid := range firstCollection {
		if !secondColSet.Exists(firstColUuid) {
			aggregatingSet.Add(firstColUuid)
		}
	}
}
