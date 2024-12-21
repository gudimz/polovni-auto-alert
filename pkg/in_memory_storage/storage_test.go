package inmemorystorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	s := New[string, int]()

	s.Set("key", 1)
	assert.Equal(t, 1, s.m["key"])
	assert.Len(t, s.m, 1)

	v, ok := s.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 1, v)

	assert.True(t, s.Contains("key"))
	assert.False(t, s.Contains("key2"))

	s.Delete("key")
	assert.Empty(t, s.m)
	assert.False(t, s.Contains("key"))
}

func TestStorage_Get(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		prepare func() *Storage[string, int]
		want    int
		exists  bool
	}{
		{
			name: "key exists",
			key:  "exist_key",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.Set("exist_key", 1)
				return s
			},
			want:   1,
			exists: true,
		},
		{
			name: "key not exist",
			key:  "not_exist_key",
			prepare: func() *Storage[string, int] {
				return New[string, int]()
			},
			exists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()

			got, exists := s.Get(tc.key)
			require.Equal(t, tc.exists, exists)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStorage_Set(t *testing.T) {
	testCases := []struct {
		name string
		args struct {
			key string
			val int
		}
		want map[string]int
	}{
		{
			name: "set key",
			args: struct {
				key string
				val int
			}{key: "key", val: 1},
			want: map[string]int{"key": 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New[string, int]()
			s.Set(tc.args.key, tc.args.val)
			assert.Equal(t, tc.want, s.m)
		})
	}
}

func TestStorage_SetBatch(t *testing.T) {
	testCases := []struct {
		name string
		args map[string]int
		want map[string]int
	}{
		{
			name: "set batch",
			args: map[string]int{"key1": 1, "key2": 2},
			want: map[string]int{"key1": 1, "key2": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New[string, int]()
			s.SetBatch(tc.args)
			assert.Equal(t, tc.want, s.m)
		})
	}
}

func TestStorage_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		key     string
		want    map[string]int
	}{
		{
			name: "delete key",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.Set("key", 1)
				return s
			},
			key:  "key",
			want: map[string]int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			s.Delete(tc.key)
			assert.Equal(t, tc.want, s.m)
		})
	}
}

func TestStorage_Keys(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		want    []string
	}{
		{
			name: "keys",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.Set("key1", 1)
				s.Set("key2", 2)
				return s
			},
			want: []string{"key1", "key2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			assert.ElementsMatch(t, tc.want, s.Keys())
		})
	}
}

func TestStorage_Values(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		want    []int
	}{
		{
			name: "values",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.Set("key1", 1)
				s.Set("key2", 2)
				return s
			},
			want: []int{1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			assert.ElementsMatch(t, tc.want, s.Values())
		})
	}
}

func TestStorage_Contains(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		key     string
		want    bool
	}{
		{
			name: "key exists",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.Set("key", 1)
				return s
			},
			key:  "key",
			want: true,
		},
		{
			name: "key not exists",
			prepare: func() *Storage[string, int] {
				return New[string, int]()
			},
			key:  "key",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			assert.Equal(t, tc.want, s.Contains(tc.key))
		})
	}
}

func TestStorage_Clear(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		want    map[string]int
	}{
		{
			name: "clear",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.SetBatch(map[string]int{"key1": 1, "key2": 2})
				return s
			},
			want: map[string]int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			s.Clear()
			assert.Equal(t, tc.want, s.m)
		})
	}
}

func TestStorage_Len(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func() *Storage[string, int]
		want    int
	}{
		{
			name: "len",
			prepare: func() *Storage[string, int] {
				s := New[string, int]()
				s.SetBatch(map[string]int{"key1": 1, "key2": 2})
				return s
			},
			want: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.prepare()
			assert.Equal(t, tc.want, s.Len())
		})
	}
}

func TestStorage_Replace(t *testing.T) {
	testCases := []struct {
		name string
		args map[string]int
		want map[string]int
	}{
		{
			name: "replace",
			args: map[string]int{"key1": 1, "key2": 2},
			want: map[string]int{"key1": 1, "key2": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New[string, int]()
			s.Replace(tc.args)
			assert.Equal(t, tc.want, s.m)
		})
	}
}

func TestStorage_Copy(t *testing.T) {
	testCases := []struct {
		name string
		args map[string]int
	}{
		{
			name: "copy",
			args: map[string]int{"key1": 1, "key2": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New[string, int]()
			s.SetBatch(tc.args)

			cs := s.Copy()
			assert.Equal(t, s.m, cs.m)
		})
	}
}

func TestStorage_CopyMap(t *testing.T) {
	testCases := []struct {
		name string
		args map[string]int
	}{
		{
			name: "copy map",
			args: map[string]int{"key1": 1, "key2": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New[string, int]()
			s.SetBatch(tc.args)

			cm := s.CopyMap()
			assert.Equal(t, s.m, cm)
		})
	}
}
