package logz

import "github.com/sirupsen/logrus"

func InfoOutput(logger *logrus.Logger, out string) {
	if out != "" {
		logger.Infof("%s", out)
	}
}

func DebugOutput(logger *logrus.Logger, out string) {
	if out != "" {
		logger.Debugf("%s", out)
	}
}
