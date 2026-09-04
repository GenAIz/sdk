package version

import (
	"fmt"

	"genaiz.com/genaiz/version/env"
)

var (
	version = "1.0.0"
	Head    = ""
)

func GetVersion() string {
	if Head != "" {
		Head = fmt.Sprintf(" (%s)", Head)
	}

	return fmt.Sprintf("%s%s%s", version, env.GetVersionTag(), Head)
}
