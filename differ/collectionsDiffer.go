package differ

import "github.com/Workiva/go-datastructures/set"

type CollectionsDiffer interface {
	Diff(incomingCollectionUuids []string, oldCollectionUuids []string) *set.Set
}

type defaultCollectionsDiffer struct {
}

func NewDefaultCollectionsDiffer() *defaultCollectionsDiffer {
	return &defaultCollectionsDiffer{}
}

func (dcd *defaultCollectionsDiffer) Diff(incomingCollectionUuids []string, oldCollectionUuids []string) *set.Set {
	diffSet := set.New()

	oneWayDiff(incomingCollectionUuids, oldCollectionUuids, diffSet)
	oneWayDiff(oldCollectionUuids, incomingCollectionUuids, diffSet)

	return diffSet
}

func oneWayDiff(firstCollection []string, secondCollection []string, setToAdd *set.Set) {
	secondColSet := set.New()
	for _, secondColUuid := range secondCollection {
		secondColSet.Add(secondColUuid)
	}

	for _, firstColUuid := range firstCollection {
		if !secondColSet.Exists(firstColUuid) {
			setToAdd.Add(firstColUuid)
		}
	}
}
