package types_convertation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMillisecondUnixTimestamp(t *testing.T) {
	testData := 1639473822

	ts, err := ParseMillisecondUnixTimestamp(testData)
	assert.NoError(t, err)
	t.Log(ts)
}
