package injector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testA struct {
	V int
}

type testB struct {
	A  *testA `inject:"a1"`
	A1 *testA `inject:"a2"`
}

func TestInjectorGet(t *testing.T) {

	a11 := &testA{1}
	a22 := &testA{2}

	require.NotEqual(t, a11, a22)

	Register("a1", a11)
	Register("a2", a22)

	var a1 *testA

	require.True(t, Get("a1", &a1))

	require.NotNil(t, a1)

	require.Equal(t, a11, a1)

	var a2 *testA

	require.True(t, Get("a2", &a2))

	require.NotNil(t, a2)

	require.Equal(t, a22, a2)

	require.NotEqual(t, a1, a2)

	require.False(t, Get("test", &a1))

	a := make([]*testA, 0)

	Find(&a)

	require.Equal(t, 2, len(a))

	b := &testB{}

	require.NoError(t, Inject(b))

	require.Equal(t, b.A, a1)
	require.Equal(t, b.A1, a2)

	println(fmt.Sprintf("%p", b.A))
	println(fmt.Sprintf("%p", b.A1))
}
