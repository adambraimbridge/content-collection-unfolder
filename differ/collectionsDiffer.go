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
	for _, incColUuid := range incomingCollectionUuids {
		found := false
		for _, oldColUuid := range oldCollectionUuids {
			if incColUuid == oldColUuid {
				found = true
				break
			}
		}
		if !found {
			diffCol[incColUuid] = false
		}
	}

	for _, oldColUuid := range oldCollectionUuids {
		found := false
		for _, incColUuid := range incomingCollectionUuids {
			if oldColUuid == incColUuid {
				found = true
				break
			}
		}
		if !found {
			diffCol[oldColUuid] = true
		}
	}

	return diffCol
}
