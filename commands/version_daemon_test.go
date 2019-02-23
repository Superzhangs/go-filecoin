package commands_test

import (
	"fmt"
	"os/exec"
	"testing"

	th "github.com/filecoin-project/go-filecoin/testhelpers"
	ast "github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Parallel()
	assert := ast.New(t)

	var gitOut []byte
	var err error
	gitArgs := []string{"rev-parse", "--verify", "HEAD"}

	if gitOut, err = exec.Command("git", gitArgs...).Output(); err != nil {
		assert.NoError(err)
	}
	commit := string(gitOut)

	d := th.NewDaemon(t).Start()
	defer d.ShutdownSuccess()

	out := d.RunSuccess("version")
	assert.Exactly(out.ReadStdout(), fmt.Sprintf("commit: %s", commit))
}
