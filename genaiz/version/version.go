package version

import (
	"fmt"

	"genaiz.com/genaiz/version/env"
)

var (
	version = "0.4.17"
	Head    = ""
)

func GetVersion() string {
	if Head != "" {
		Head = fmt.Sprintf(" (%s)", Head)
	}

	return fmt.Sprintf("%s%s%s", version, env.GetVersionTag(), Head)
}
