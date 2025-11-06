package version

import (
	"fmt"

	"genaiz.com/genaiz/version/env"
)

var version = "0.1.32"

func GetVersion() string {
	return fmt.Sprintf("%s%s", version, env.GetVersionTag())
}
