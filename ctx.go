package mfctx

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/myfantasy/ints"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Crumps struct {
	ctx context.Context

	mx sync.Mutex

	segment string
	method  string

	values map[string]Values

	lp LogProvider
	mp MetricsProvider
	jp JsonifyProvider
	tp TraceProvider

	operationID string
	traceID     string
	span        trace.Span
	startTime   time.Time

	steps []StepInfo
}

type StepInfo struct {
	StepID  string    `json:"step_id"`
	StartTS time.Time `json:"start"`
	Segment string    `json:"segment"`
	Method  string    `json:"method"`
}
type Values []json.RawMessage

func valuesMapCopy(vsm map[string]Values) map[string]Values {
	res := make(map[string]Values, len(vsm))

	for k, v := range vsm {
		res[k] = valuesCopy(v)
	}

	return res
}

var _ context.Context = &Crumps{}

func valuesCopy(vs Values) Values {
	res := make(Values, 0, len(vs))

	if len(vs) >= 0 {
		res = append(res, vs...)
	}

	return res
}

func stepsCopy(s []StepInfo) []StepInfo {
	res := make([]StepInfo, 0, len(s)+1)

	res = append(res, s...)

	return res
}

func (c *Crumps) Copy() *Crumps {
	if c == nil || c.ctx == nil {
		return &Crumps{
			ctx:    context.Background(),
			values: valuesMapCopy(nil),

			operationID: ints.DefaultUuidGenerator.Next().String(),
			steps:       make([]StepInfo, 0, 1),
		}
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	return &Crumps{
		ctx:     c.ctx,
		values:  valuesMapCopy(c.values),
		method:  c.method,
		segment: c.segment,

		operationID: c.operationID,
		traceID:     c.traceID,

		lp: c.lp,
		mp: c.mp,
		jp: c.jp,
		tp: c.tp,

		steps: stepsCopy(c.steps),
	}
}

func FromCtx(ctx context.Context) *Crumps {
	c, _ := ctx.Value(CtxValName).(*Crumps)

	c = c.Copy()

	c.ctx = context.WithValue(ctx, CtxValName, c)

	return c
}

func (c *Crumps) Deadline() (deadline time.Time, ok bool) {
	if c == nil || c.ctx == nil {
		return deadline, ok
	}

	return c.ctx.Deadline()
}

func (c *Crumps) Done() <-chan struct{} {
	if c == nil || c.ctx == nil {
		return make(chan struct{})
	}

	return c.ctx.Done()
}

func (c *Crumps) Err() error {
	if c == nil || c.ctx == nil {
		return nil
	}

	return c.ctx.Err()
}

func (c *Crumps) Value(key any) any {
	if key == CtxValName {
		return c
	}

	if c == nil || c.ctx == nil {
		return nil
	}

	return c.ctx.Value(key)
}

func (c *Crumps) methodFullName() string {
	if c == nil || c.ctx == nil {
		return ""
	}

	return c.segment + "." + c.method
}

func (c *Crumps) StartSegment(segment, method string) *Crumps {
	res := c.Copy()

	res.segment = segment
	res.method = method

	startTime := time.Now()
	WriteMetricRequest(res.mp, res, segment, method)

	ctx, span := otel.Tracer(appName).Start(res.ctx, res.methodFullName())
	res.traceID = span.SpanContext().TraceID().String()
	res.ctx = ctx
	res.startTime = startTime
	res.span = span
	res.steps = append(res.steps,
		StepInfo{
			StepID:  ints.DefaultUuidGenerator.Next().String(),
			StartTS: startTime,
			Segment: segment,
			Method:  method,
		},
	)

	res.Log(Start, "")

	return res
}

func (c *Crumps) Start(method string) *Crumps {
	return c.StartSegment(c.segment, method)
}

func (c *Crumps) Complete(err error) {
	if c.span != nil {
		WriteTrace(c.tp, c.span, c, c.segment, c.method, c.operationID, c.steps, valuesMapCopy(c.values))

		if err != nil {
			//fp.span.SetAttributes(attribute.String("operationResult", MsgError))
			c.span.SetStatus(codes.Error, c.methodFullName()+" FAIL")
			c.span.RecordError(err)

			WriteMetricResponse(c.mp, c, c.startTime, c.segment, c.method, MsgError)
			c.Log(FinishError, err.Error())
		} else {
			//fp.span.SetAttributes(attribute.String("operationResult", MsgComplete))

			WriteMetricResponse(c.mp, c, c.startTime, c.segment, c.method, MsgComplete)
			c.Log(Finish, "")
		}

		c.span.End()
	}
}

// With adds jsonify value and adds it
func (c *Crumps) With(name string, value any) *Crumps {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.values[name] = append(c.values[name], ToJson(c.jp, value))

	return c
}

func (c *Crumps) Log(level LogLevel, message string) {
	if c == nil {
		return
	}

	WriteLog(c.lp, c, level, message,
		c.segment, c.method, c.operationID, c.traceID, c.steps, valuesMapCopy(c.values))
}

func (c *Crumps) GetOperationID() string {
	if c == nil {
		return ""
	}
	return c.operationID
}

func (c *Crumps) WithCancelCause() (*Crumps, context.CancelCauseFunc) {
	c = c.Copy()
	var cancel context.CancelCauseFunc
	c.ctx, cancel = context.WithCancelCause(c.ctx)

	return c, cancel
}

func (c *Crumps) WithCancel() (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithCancel(c.ctx)

	return c, cancel
}

func (c *Crumps) WithTimeoutCause(timeout time.Duration, cause error) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithTimeoutCause(c.ctx, timeout, cause)

	return c, cancel
}

func (c *Crumps) WithTimeout(timeout time.Duration) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithTimeout(c.ctx, timeout)

	return c, cancel
}

func (c *Crumps) WithDeadlineCause(d time.Time, cause error) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithDeadlineCause(c.ctx, d, cause)

	return c, cancel
}

func (c *Crumps) WithDeadline(d time.Time) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithDeadline(c.ctx, d)

	return c, cancel
}

func (c *Crumps) WithValue(key any, val any) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx = context.WithValue(c.ctx, key, val)

	return c, cancel
}

func (c *Crumps) WithoutCancel(key any, val any) (*Crumps, context.CancelFunc) {
	c = c.Copy()
	var cancel context.CancelFunc
	c.ctx = context.WithoutCancel(c.ctx)

	return c, cancel
}
