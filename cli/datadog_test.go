package cli

import (
	"os/user"
	"testing"
)

func TestExpandPath(t *testing.T) {
	path1 := expandPath("~/.datadog")

	usr, _ := user.Current()
	dir := usr.HomeDir

	if path1 != dir+"/.datadog" {
		t.Errorf("expandPath(\"~/.datadog\") = %s", path1)
	}
}
