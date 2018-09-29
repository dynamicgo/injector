package injector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testA struct {
}

type testB struct {
	A  *testA `inject:"a1"`
	A1 *testA `inject:"a2"`
}

func TestInjectorGet(t *testing.T) {
	Register("a1", &testA{})
	Register("a2", &testA{})

	var a1 *testA

	require.True(t, Get("a1", &a1))

	var a2 *testA

	require.True(t, Get("a2", &a2))

	require.False(t, Get("test", &a1))

	a := make([]*testA, 0)

	Find(&a)

	require.Equal(t, 2, len(a))

	b := &testB{}

	require.NoError(t, Inject(b))

	require.Equal(t, b.A, a1)
	require.Equal(t, b.A, a2)
}
