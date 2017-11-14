package differ

type CollectionsDiffer interface {
	Diff(incomingCollectionUuids []string, oldCollectionUuids []string) ([]string, map[string]bool)
}

type defaultCollectionsDiffer struct {
}

func NewDefaultCollectionsDiffer() *defaultCollectionsDiffer {
	return &defaultCollectionsDiffer{}
}

func (dcd *defaultCollectionsDiffer) Diff(incomingCollectionUuids []string, oldCollectionUuids []string) ([]string, map[string]bool) {
	var diffColUuids []string
	isDeleted := make(map[string]bool)
	for _, incColUuid := range incomingCollectionUuids {
		found := false
		for _, oldColUuid := range oldCollectionUuids {
			if incColUuid == oldColUuid {
				found = true
				break
			}
		}
		if !found {
			diffColUuids = append(diffColUuids, incColUuid)
			isDeleted[incColUuid] = false
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
			diffColUuids = append(diffColUuids, oldColUuid)
			isDeleted[oldColUuid] = true
		}
	}

	return diffColUuids, isDeleted
}
