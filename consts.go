package mfctx

import "github.com/myfantasy/ints"

type LogLevel string

const (
	CtxValName = "_Crumps"
)

const (
	Trace       LogLevel = "trace"
	Debug       LogLevel = "debug"
	Start       LogLevel = "start"
	Finish      LogLevel = "finish"
	Info        LogLevel = "info"
	FinishError LogLevel = "finish_error"
	Warning     LogLevel = "warning"
	Error       LogLevel = "error"
	Fatal       LogLevel = "fatal"
)

const (
	MsgComplete = "complete"
	MsgError    = "error"
)

var appID = ints.DefaultUuidGenerator.Next().String()

var (
	appName    = "appName"
	appVersion = "appVersion"
)

func SetAppName(name string) {
	appName = name
}

func GetAppName() (name string) {
	return appName
}

func SetAppVersion(version string) {
	appVersion = version
}

func GetAppVersion() (version string) {
	return appVersion
}

func GetAppID() string {
	return appID
}
