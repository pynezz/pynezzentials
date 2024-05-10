/*

This package contains functions that interact with the system directly.
It is not recommended to use this package if you're not comfortable with that.

Due to the possible security issues, this will not be dependent on the other packages in this module.

*/

package cmds

import (
	"fmt"
	"os/exec"
)

// GitCommits returns the latest git commit number
// Minor warning: This will perform a command directly on the system. If you're not comfortable with that, you should not import the /cmds package
func GitCommits() string {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	version, err := cmd.Output()
	if err != nil {
		errorf("failed to get git commit count: %v", err)
	}
	return fmt.Sprintf("%s\n", version)
}

// ErrorF returns a formatted error message
func errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s[!]%s %s", red, reset, fmt.Sprintf(format, a...))
}
