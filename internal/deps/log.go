package deps

import "context"

// Log is the logging surface used by the internal packages. It intentionally
// re-declares only the methods the domain needs so that internal code does not
// depend on the concrete logger implementation in pkg/clog. The production
// logger (clog.CLog) satisfies this interface structurally.
type Log interface {
	InfoCtx(ctx context.Context, msg string, args ...any)
	DebugCtx(ctx context.Context, msg string, args ...any)
	WarnCtx(ctx context.Context, msg string, args ...any)
	ErrorCtx(ctx context.Context, err error, msg string, args ...any)
}

// LogStub is a no-op Log for use in unit tests.
type LogStub struct{}

func NewLogStub() *LogStub { return &LogStub{} }

func (s *LogStub) InfoCtx(context.Context, string, ...any)         {}
func (s *LogStub) DebugCtx(context.Context, string, ...any)        {}
func (s *LogStub) WarnCtx(context.Context, string, ...any)         {}
func (s *LogStub) ErrorCtx(context.Context, error, string, ...any) {}
