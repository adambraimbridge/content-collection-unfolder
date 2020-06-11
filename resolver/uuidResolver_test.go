package resolver

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidInput(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection.json")

	r := NewUuidResolver()
	uuidsAndDate, err := r.Resolve(ccBytes)

	assert.NoError(t, err)
	assert.Equal(t, "2017-01-31T15:33:21.687Z", uuidsAndDate.LastModified)
	assert.Equal(t, 3, len(uuidsAndDate.UuidArr))
	assert.Contains(t, uuidsAndDate.UuidArr, "aaaac4c6-dcc6-11e6-86ac-f253db7791c6")
	assert.Contains(t, uuidsAndDate.UuidArr, "bbbbc4c6-dcc6-11e6-86ac-f253db7791c6")
	assert.Contains(t, uuidsAndDate.UuidArr, "d4986a58-de3b-11e6-86ac-f253db7791c6")
}

func TestEmptyItems(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-empty-items.json")

	r := NewUuidResolver()
	uuidsAndDate, err := r.Resolve(ccBytes)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(uuidsAndDate.UuidArr))
}

func TestNoItems(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-no-items.json")

	r := NewUuidResolver()
	uuidsAndDate, err := r.Resolve(ccBytes)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(uuidsAndDate.UuidArr))
}

func TestNoLastModified(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-no-lastModified.json")

	r := NewUuidResolver()
	_, err := r.Resolve(ccBytes)

	assert.Error(t, err)
}

func TestNoUuid(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-no-uuid.json")

	r := NewUuidResolver()
	_, err := r.Resolve(ccBytes)

	assert.Error(t, err)
}

func TestInvalidLastModified(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-invalid-lastModified.json")

	r := NewUuidResolver()
	_, err := r.Resolve(ccBytes)

	assert.Error(t, err)
}

func TestInvalidUUID(t *testing.T) {
	ccBytes := readTestFile(t, "content-collection-invalid-uuid.json")

	r := NewUuidResolver()
	_, err := r.Resolve(ccBytes)

	assert.Error(t, err)
}

func readTestFile(t *testing.T, fileName string) []byte {
	file, err := os.Open("../test-resources/" + fileName)
	assert.NoError(t, err)

	defer file.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	assert.NoError(t, err)

	return buf.Bytes()
}
