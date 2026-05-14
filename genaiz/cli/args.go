package cli

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func ArgsHostAndPort(address string) (string, int, error) {
	var host, portString string
	var err error

	if address == "*" {
		return address, 0, nil
	} else if host, portString, err = net.SplitHostPort(address); err == nil {
		var port int

		if port, err = strconv.Atoi(portString); err == nil {
			if port > 0 && port <= 65535 {
				return host, port, nil
			}

			return "", -1, net.InvalidAddrError("invalid port")
		}
	}

	return "", -1, err
}

func ArgsOptionalFolder(typeName string, maxSize int, validates config.Validates) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) >= maxSize {
			var lastArg = maxSize - 1

			if info, err := os.Stat(args[lastArg]); err == nil {
				if !info.IsDir() {
					return fmt.Errorf("%s is not a folder", args[lastArg])
				}
			} else if !validates(filepath.Base(args[lastArg])) {
				return fmt.Errorf("%s is not a valid %s name", typeName, args[lastArg])
			}
		}

		return nil
	}
}

func ArgsOptionalSingle(args []string) string {
	if len(args) == 1 {
		return args[0]
	}

	return strings.Join(args, " ")
}
