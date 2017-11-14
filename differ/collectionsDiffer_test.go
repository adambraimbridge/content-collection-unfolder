package differ

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectionDiffer_Diff_Ok(t *testing.T) {
	collectionsDiffer := NewDefaultCollectionsDiffer()

	incomingCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "077f67ef-e827-49f8-8207-01c7720cbd53", "79b5a80e-96a7-4ac8-b168-5406910de419"}
	oldCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}

	expectedDiff := []string{"077f67ef-e827-49f8-8207-01c7720cbd53", "79b5a80e-96a7-4ac8-b168-5406910de419", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}
	expectedIsDeletedMap := map[string]bool{"077f67ef-e827-49f8-8207-01c7720cbd53": false, "79b5a80e-96a7-4ac8-b168-5406910de419": false, "ef0d9b7f-c3e9-4692-9e62-1a38789af24a": true}

	actualDiffCol, actualIsDeletedMap := collectionsDiffer.Diff(incomingCol, oldCol)

	assert.Equal(t, expectedDiff, actualDiffCol)
	assert.Equal(t, expectedIsDeletedMap, actualIsDeletedMap)
}

func TestCollectionDiffer_Diff_EmptyIncomingCol_Ok(t *testing.T) {
	collectionsDiffer := NewDefaultCollectionsDiffer()

	var incomingCol []string
	oldCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}

	expectedDiff := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}
	expectedIsDeletedMap := map[string]bool{"9e917253-10d2-46d8-ab3b-b510dc3a7abf": true, "ef0d9b7f-c3e9-4692-9e62-1a38789af24a": true}

	actualDiffCol, actualIsDeletedMap := collectionsDiffer.Diff(incomingCol, oldCol)

	assert.Equal(t, expectedDiff, actualDiffCol)
	assert.Equal(t, expectedIsDeletedMap, actualIsDeletedMap)
}

func TestCollectionDiffer_Diff_EmptyOldCol_Ok(t *testing.T) {
	collectionsDiffer := NewDefaultCollectionsDiffer()

	incomingCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}
	var oldCol []string

	expectedDiff := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}
	expectedIsDeletedMap := map[string]bool{"9e917253-10d2-46d8-ab3b-b510dc3a7abf": false, "ef0d9b7f-c3e9-4692-9e62-1a38789af24a": false}

	actualDiffCol, actualIsDeletedMap := collectionsDiffer.Diff(incomingCol, oldCol)

	assert.Equal(t, expectedDiff, actualDiffCol)
	assert.Equal(t, expectedIsDeletedMap, actualIsDeletedMap)
}

func TestCollectionDiffer_Diff_SameCollections_Ok(t *testing.T) {
	collectionsDiffer := NewDefaultCollectionsDiffer()

	incomingCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}
	oldCol := []string{"9e917253-10d2-46d8-ab3b-b510dc3a7abf", "ef0d9b7f-c3e9-4692-9e62-1a38789af24a"}

	var expectedDiff []string
	expectedIsDeletedMap := map[string]bool{}

	actualDiffCol, actualIsDeletedMap := collectionsDiffer.Diff(incomingCol, oldCol)

	assert.Equal(t, expectedDiff, actualDiffCol)
	assert.Equal(t, expectedIsDeletedMap, actualIsDeletedMap)
}

func TestCollectionDiffer_Diff_EmptyCollections_Ok(t *testing.T) {
	collectionsDiffer := NewDefaultCollectionsDiffer()

	var incomingCol []string
	var oldCol []string

	var expectedDiff []string
	expectedIsDeletedMap := map[string]bool{}

	actualDiffCol, actualIsDeletedMap := collectionsDiffer.Diff(incomingCol, oldCol)

	assert.Equal(t, expectedDiff, actualDiffCol)
	assert.Equal(t, expectedIsDeletedMap, actualIsDeletedMap)
}
