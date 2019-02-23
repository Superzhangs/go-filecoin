package commands_test

import (
	"testing"

	th "github.com/filecoin-project/go-filecoin/testhelpers"

	ast "github.com/stretchr/testify/assert"
)

func TestBootstrapList(t *testing.T) {
	t.Parallel()
	assert := ast.New(t)

	d := th.NewDaemon(t).Start()
	defer d.ShutdownSuccess()

	bs := d.RunSuccess("bootstrap ls")

	assert.Equal("&{[]}\n", bs.ReadStdout())
}
