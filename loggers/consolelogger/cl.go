package consolelogger

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/myfantasy/mfctx"
	"github.com/myfantasy/mfctx/jsonify"
)

const (
	TSName  = "ts"
	LVName  = "level"
	MSGName = "msg"
	SGName  = "segment"
	MTName  = "method"
	OPName  = "op_id"
	TRName  = "tr_id"

	AppID   = "app_id"
	AppName = "app_n"
	Version = "version"
	DCName  = "dc"

	STName = "steps"
)

var Order map[string]int = map[string]int{
	TSName:  -99,
	LVName:  -98,
	MSGName: -97,
	SGName:  -96,
	MTName:  -95,
	OPName:  -94,
	TRName:  -93,

	STName: -80,
}

var DefaultAllowLevels map[mfctx.LogLevel]bool = map[mfctx.LogLevel]bool{
	mfctx.Trace:       false,
	mfctx.Debug:       false,
	mfctx.Start:       true,
	mfctx.Finish:      true,
	mfctx.Info:        true,
	mfctx.FinishError: true,
	mfctx.Warning:     true,
	mfctx.Error:       true,
	mfctx.Fatal:       true,
}

type SimpleConsoleLogger struct {
	AllowLevels map[mfctx.LogLevel]bool
}

type valType struct {
	name  string
	value json.RawMessage
}

type vals []valType

func (l *SimpleConsoleLogger) WriteLog(
	_ *mfctx.Crumps, level mfctx.LogLevel, message, segment, method, operationID, traceID string, steps []mfctx.StepInfo, values map[string]mfctx.Values) {
	if l.AllowLevels == nil {
		l.AllowLevels = DefaultAllowLevels
	}

	if !l.AllowLevels[level] {
		return
	}

	vls := make(vals, 0)
	t := time.Now().Truncate(time.Microsecond)

	vls = append(vls, valType{name: TSName, value: jsonify.JsonifyM(t)})
	vls = append(vls, valType{name: LVName, value: jsonify.JsonifyM(level)})

	if len(message) > 0 {
		vls = append(vls, valType{name: MSGName, value: jsonify.JsonifyM(message)})
	}
	if len(segment) > 0 {
		vls = append(vls, valType{name: SGName, value: jsonify.JsonifyM(segment)})
	}
	if len(method) > 0 {
		vls = append(vls, valType{name: MTName, value: jsonify.JsonifyM(method)})
	}
	vls = append(vls, valType{name: OPName, value: jsonify.JsonifyM(operationID)})
	vls = append(vls, valType{name: TRName, value: jsonify.JsonifyM(traceID)})

	vls = append(vls, valType{name: AppName, value: jsonify.JsonifyM(mfctx.GetAppName())})
	vls = append(vls, valType{name: AppID, value: jsonify.JsonifyM(mfctx.GetAppID())})
	vls = append(vls, valType{name: Version, value: jsonify.JsonifyM(mfctx.GetAppVersion())})
	vls = append(vls, valType{name: DCName, value: jsonify.JsonifyM(mfctx.GetDataCenter())})

	if len(steps) > 0 {
		vls = append(vls, valType{name: STName, value: jsonify.JsonifyM(steps)})
	}

	for k, v := range values {
		if len(v) == 1 {
			vls = append(vls, valType{name: k, value: v[0]})
		} else {
			vls = append(vls, valType{name: k, value: jsonify.JsonifyM(v)})
		}
	}

	sort.Slice(vls, func(i, j int) bool {
		if Order[vls[i].name] > Order[vls[j].name] {
			return false
		}
		if Order[vls[i].name] < Order[vls[j].name] {
			return true
		}

		return vls[i].name < vls[j].name
	})

	result := []byte("{")
	for i, v := range vls {
		if i > 0 {
			result = append(result, []byte(",")...)
		}
		result = append(result, []byte("\""+v.name+"\":")...)
		result = append(result, v.value...)
	}
	result = append(result, []byte("}\n")...)

	os.Stderr.Write(result)
}
