package slices

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	s := []string{"hello", "world"}

	r := Map(s, func(s string) int {
		return len(s)
	})

	assert.Equal(t, []int{5, 5}, r)
}

func TestTryMap(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		s := []string{"hello", "world"}

		r, err := TryMap(s, func(s string) (int, error) {
			return len(s), nil
		})
		require.NoError(t, err)

		assert.Equal(t, []int{5, 5}, r)
	})

	t.Run("Error", func(t *testing.T) {
		s := []string{"hello", "world"}

		expectedErr := errors.New("something went wrong")

		r, err := TryMap(s, func(s string) (int, error) {
			return len(s), expectedErr
		})
		require.Error(t, err)

		assert.Nil(t, r)
		assert.Equal(t, expectedErr, err)
	})
}
