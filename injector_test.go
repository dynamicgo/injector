package injector

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type testA struct {
}

type testB struct {
	a  testA `inject:"nullable"`
	a1 testA `inject:""`
}

func TestStructExecutor(t *testing.T) {
	executor, err := NewStructExecutor(reflect.TypeOf(&testB{}))

	require.NoError(t, err)

	require.Equal(t, 2, len(executor.Ops))

	op := executor.Ops[0].(*injectFieldOp)

	require.True(t, op.nullable)

	op = executor.Ops[1].(*injectFieldOp)

	require.False(t, op.nullable)
}

func BenchmarkCreateStructExecutor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewStructExecutor(reflect.TypeOf(&testB{}))
	}
}
