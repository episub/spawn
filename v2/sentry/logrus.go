package sentry

import (
	"fmt"

	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

// LogrusHook Hook for sending logrus events to Sentry.  Assumes that sentry
// has already been configured.
type LogrusHook struct {
	ActiveLevels []logrus.Level
}

// Levels Return which levels we are serving
func (lh LogrusHook) Levels() []logrus.Level {
	return lh.ActiveLevels
	//	return []logrus.Level{
	//		logrus.ErrorLevel,
	//		logrus.FatalLevel,
	//		logrus.PanicLevel,
	//	}
}

// Fire Called when logrus sends us an event
func (lh LogrusHook) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil
	}

	e := *entry

	sentry.WithScope(func(scope *sentry.Scope) {
		switch e.Level {
		case logrus.TraceLevel, logrus.DebugLevel:
			scope.SetLevel(sentry.LevelDebug)
		case logrus.InfoLevel:
			scope.SetLevel(sentry.LevelInfo)
		case logrus.WarnLevel:
			scope.SetLevel(sentry.LevelWarning)
		case logrus.ErrorLevel:
			scope.SetLevel(sentry.LevelError)
		case logrus.FatalLevel, logrus.PanicLevel:
			scope.SetLevel(sentry.LevelFatal)
		}

		// Add fields as context (not tags)
		for k, v := range e.Data {
			scope.SetContext(k, v)
		}
		sentry.CaptureException(fmt.Errorf(e.Message))
	})

	return nil
}
