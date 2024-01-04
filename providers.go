package mfctx

import (
	"encoding/json"
	"time"

	"github.com/myfantasy/mfctx/jsonify"
	"go.opentelemetry.io/otel/trace"
)

type LogProvider interface {
	WriteLog(c *Crumps, level LogLevel, message, segment, method, operationID, traceID string, steps []StepInfo, values map[string]Values)
}

type TraceProvider interface {
	WriteTrace(span trace.Span, c *Crumps, segment, method, operationID string, steps []StepInfo, values map[string]Values)
}

type MetricsProvider interface {
	WriteMetricRequest(c *Crumps, segment, method string)
	WriteMetricResponse(c *Crumps, mRequest time.Time, segment, method string, resultResult string)
}

type JsonifyProvider interface {
	ToJson(value any) json.RawMessage
}

type Provider struct {
	LP LogProvider
	MP MetricsProvider
	SP JsonifyProvider
	TP TraceProvider
}

var DefaultProvider *Provider = nil

func SetDefaultProvider(p *Provider) {
	DefaultProvider = p
}

func (p *Provider) MakeCrumps() (c *Crumps) {
	c = c.Copy()

	c.lp = p.LP
	c.mp = p.MP

	return c
}

func WriteMetricRequest(mp MetricsProvider, c *Crumps, segment, method string) {
	if mp == nil {
		if DefaultProvider != nil {
			mp = DefaultProvider.MP
		}
	}

	if mp != nil {
		mp.WriteMetricRequest(c, segment, method)
	}
}
func WriteMetricResponse(mp MetricsProvider, c *Crumps, mRequest time.Time, segment, method string, resultResult string) {
	if mp == nil {
		if DefaultProvider != nil {
			mp = DefaultProvider.MP
		}
	}

	if mp != nil {
		mp.WriteMetricResponse(c, mRequest, segment, method, resultResult)
	}
}

func WriteLog(lp LogProvider, c *Crumps, level LogLevel, message, segment, method, operationID, traceID string, steps []StepInfo, values map[string]Values) {
	if lp == nil {
		if DefaultProvider != nil {
			lp = DefaultProvider.LP
		}
	}

	if lp != nil {
		lp.WriteLog(c, level, message, segment, method, operationID, traceID, steps, values)
	}
}

func ToJson(sp JsonifyProvider, value any) json.RawMessage {
	if sp == nil {
		if DefaultProvider != nil {
			sp = DefaultProvider.SP
		}
	}

	if sp == nil {
		return jsonify.Jsonify(value, "", "")
	}

	return sp.ToJson(value)
}

func WriteTrace(tp TraceProvider, span trace.Span, c *Crumps, segment, method, operationID string, steps []StepInfo, values map[string]Values) {
	if tp == nil {
		if DefaultProvider != nil {
			tp = DefaultProvider.TP
		}
	}

	if tp != nil {
		tp.WriteTrace(span, c, segment, method, operationID, steps, values)
	}
}
