package commontracer

import (
	"github.com/myfantasy/mfctx"
	"github.com/myfantasy/mfctx/jsonify"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	SGName = "segment"
	MTName = "method"
	OPName = "op_id"

	STName = "steps"
)

type SimpleTraicer struct {
}

func (l *SimpleTraicer) WriteTrace(
	span trace.Span, c *mfctx.Crumps, segment, method, operationID string, steps []mfctx.StepInfo, values map[string]mfctx.Values) {

	if len(segment) > 0 {
		span.SetAttributes(attribute.String(SGName, segment))
	}
	if len(method) > 0 {
		span.SetAttributes(attribute.String(MTName, method))
	}
	span.SetAttributes(attribute.String(OPName, operationID))

	if len(steps) > 0 {
		span.SetAttributes(attribute.String(STName, string(jsonify.JsonifyM(steps))))
	}

	for k, v := range values {
		if len(v) == 1 {
			span.SetAttributes(attribute.String(k, string(v[0])))
		} else {
			span.SetAttributes(attribute.String(k, string(jsonify.JsonifyM(v))))
		}
	}
}
