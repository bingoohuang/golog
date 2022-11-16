package logfmt

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeLayout(t *testing.T) {
	timePart, err := parseTime(true, "", "yyyy-MM-dd HH:mm:ss,SSS")
	assert.Nil(t, err)
	var b bytes.Buffer
	timePart.Append(&b, EntryItem{
		EntryTime: time.Now(),
	})
	t.Log(b.String())
}
