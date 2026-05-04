package version

import (
	"fmt"

	"genaiz.com/genaiz/version/env"
)

var version = "0.3.4"

func GetVersion() string {
	return fmt.Sprintf("%s%s", version, env.GetVersionTag())
}
