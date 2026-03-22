package clog

import (
	"context"
)

const FakeCLogStubField = "abc"

func NewCLogStub() *Stub {
	return &Stub{}
}

type Stub struct{}

func (l *Stub) ErrorCtx(_ context.Context, _ error, _ string, _ ...any) {
}

func (l *Stub) InfoCtx(_ context.Context, _ string, _ ...any) {
}

func (l *Stub) DebugCtx(_ context.Context, _ string, _ ...any) {
}

func (l *Stub) WarnCtx(_ context.Context, _ string, _ ...any) {
}

func (l *Stub) AddKeysValuesToCtx(ctx context.Context, _ map[string]any) context.Context {
	return ctx
}

func (l *Stub) GetFieldByKey(_ context.Context, _ string) (any, bool) {
	return FakeCLogStubField, true
}
