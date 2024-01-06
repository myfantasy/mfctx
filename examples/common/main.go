package main

import (
	"context"
	"fmt"
	"log"

	"github.com/myfantasy/mfctx"
	"github.com/myfantasy/mfctx/loggers/consolelogger"
	"github.com/myfantasy/mfctx/metrics/metricscommon"
	"github.com/myfantasy/mfctx/tracers/commontracer"

	"github.com/myfantasy/mfctx/tracers/traceinit"
)

func main() {
	mfctx.SetAppName("mfctx.common.example")
	mfctx.SetAppVersion("draft")

	tp, err := traceinit.NewConsoleTracer()

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	mfctx.DefaultProvider = &mfctx.Provider{
		LP: &consolelogger.SimpleConsoleLogger{},
		//TP: &commontracer.SimpleTraicer{},
		TP: &commontracer.SimpleTraicer{},
		MP: metricscommon.NewMetricsCommon().AutoRegister(),
	}

	f1(context.Background())
}

func f1(ctxin context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxin).StartSegment("sg", "f1")
	defer func() {
		ctx.Complete(err)
	}()
	f2(ctx)

	return nil
}

func f2(ctxin context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxin).StartSegment("sg0", "f2")
	defer func() {
		ctx.Complete(err)
	}()

	ctx.With("as", 0).With("bla", "1")

	f3(ctx)

	ctx.With("as", 1).With("ttt", "3")

	return fmt.Errorf("errrr")
}

func f3(ctxin context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxin).StartSegment("sg1", "f3")
	defer func() {
		ctx.Complete(err)
	}()

	return nil
}
