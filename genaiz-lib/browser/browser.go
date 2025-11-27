package browser

import (
	"io"
	"os/exec"
)

var (
	shellProviders = []string{"xdg-open", "x-www-browser", "www-browser", "open"}
)

func OpenUrl(url string, errOut io.Writer, out io.Writer) error {
	return openUrl(url, errOut, out, shellProviders...)
}

func openUrl(url string, errOut io.Writer, out io.Writer, providers ...string) error {
	for _, provider := range providers {
		if _, err := exec.LookPath(provider); err == nil {
			var cmd = exec.Command(provider, url)

			cmd.Stdout = out
			cmd.Stderr = errOut
			return cmd.Run()
		}
	}

	return exec.ErrNotFound
}
