package differ

type CollectionsDiffer interface {
	Diff(incomingCollectionUuids []string, oldCollectionUuids []string) *Set
}

type defaultCollectionsDiffer struct {
}

func NewDefaultCollectionsDiffer() *defaultCollectionsDiffer {
	return &defaultCollectionsDiffer{}
}

func (dcd *defaultCollectionsDiffer) Diff(incomingCollectionUuids []string, oldCollectionUuids []string) *Set {
	diffSet := NewSet()

	oneWayDiff(incomingCollectionUuids, oldCollectionUuids, diffSet)
	oneWayDiff(oldCollectionUuids, incomingCollectionUuids, diffSet)

	return diffSet
}

func oneWayDiff(firstCollection []string, secondCollection []string, setToAdd *Set) {
	secondColSet := NewSet()
	for _, secondColUuid := range secondCollection {
		secondColSet.Add(secondColUuid)
	}

	for _, firstColUuid := range firstCollection {
		if !secondColSet.Contains(firstColUuid) {
			setToAdd.Add(firstColUuid)
		}
	}
}
