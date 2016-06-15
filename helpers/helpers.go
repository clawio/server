package helpers

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/clawio/clawiod/config"

	"github.com/Sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SanitizeURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}
	copy := *uri
	params := copy.Query()
	if len(params.Get("access_token")) > 0 {
		params.Set("access_token", "REDACTED")
		copy.RawQuery = params.Encode()
	}
	return copy.String()
}
func RedactString(v string) string {
	length := len(v)
	if length == 0 {
		return ""
	}
	if length == 1 {
		return "X"
	}
	half := length / 2
	right := v[half:]
	hidden := strings.Repeat("X", 10)
	return strings.Join([]string{hidden, right}, "")
}

func GetAppLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return NewLogger(dirs.Server.AppLogLevel, dirs.Server.AppLog,
		dirs.Server.AppLogMaxSize, dirs.Server.AppLogMaxAge, dirs.Server.AppLogMaxBackups)
}

func GetHTTPAccessLogger(conf *config.Config) *logrus.Entry {
	dirs := conf.GetDirectives()
	return NewLogger(dirs.Server.HTTPAccessLogLevel, dirs.Server.HTTPAccessLog,
		dirs.Server.HTTPAccessLogMaxSize, dirs.Server.HTTPAccessLogMaxAge, dirs.Server.HTTPAccessLogMaxBackups)

}

func NewLogger(level, writer string, maxSize, maxAge, maxBackups int) *logrus.Entry {
	base := logrus.New()

	switch writer {
	case "stdout":
		base.Out = os.Stdout
	case "stderr":
		base.Out = os.Stderr
	case "":
		base.Out = ioutil.Discard
	default:
		base.Out = &lumberjack.Logger{
			Filename:   writer,
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: maxBackups,
		}
	}

	logrusLevel, err := logrus.ParseLevel(level)
	// if provided level is not supported, default to Info level
	if err != nil {
		base.Error(err)
		logrusLevel = logrus.InfoLevel
	}
	base.Level = logrusLevel

	log := logrus.NewEntry(base)
	return log
}
