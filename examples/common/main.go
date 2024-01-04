package main

import (
	"context"
	"fmt"

	"github.com/myfantasy/mfctx"
	"github.com/myfantasy/mfctx/loggers/consolelogger"
	"github.com/myfantasy/mfctx/tracers/commontracer"
)

func main() {
	mfctx.DefaultProvider = &mfctx.Provider{
		LP: &consolelogger.SimpleConsoleLogger{},
		//TP: &commontracer.SimpleTraicer{},
		TP: &commontracer.SimpleTraicer{},
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

	return fmt.Errorf("errrr")
}
