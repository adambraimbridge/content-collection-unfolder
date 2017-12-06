package differ

type CollectionsDiffer interface {
	Diff(incomingCollectionUuids []string, oldCollectionUuids []string) (map[string]bool)
}

type defaultCollectionsDiffer struct {
}

func NewDefaultCollectionsDiffer() *defaultCollectionsDiffer {
	return &defaultCollectionsDiffer{}
}

func (dcd *defaultCollectionsDiffer) Diff(incomingCollectionUuids []string, oldCollectionUuids []string) (map[string]bool) {
	diffCol := make(map[string]bool)

	oneWayDiff(incomingCollectionUuids, oldCollectionUuids, false, diffCol)
	oneWayDiff(oldCollectionUuids, incomingCollectionUuids, true, diffCol)

	return diffCol
}

func oneWayDiff(firstCollection []string, secondCollection []string, markDeleted bool, mapToAdd map[string]bool) {
	secondCollectionTemp := make(map[string]struct{})
	var exists = struct{}{}
	for _, secondColUuid := range secondCollection {
		secondCollectionTemp[secondColUuid] = exists
	}

	for _, firstColUuid := range firstCollection {
		if _, ok := secondCollectionTemp[firstColUuid]; !ok {
			mapToAdd[firstColUuid] = markDeleted
		}
	}
}
