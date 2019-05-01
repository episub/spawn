package opa

import (
	"context"
	"testing"
)

func TestGetCompilerEmpty(t *testing.T) {
	if GetCompiler(context.Background()) == nil {
		t.Error("Expected compiler, but it was not nil")
	}
}
