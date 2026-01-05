package cache

import (
	"reflect"
	"testing"
)

func setupCache() *Cache[string, string] {
	c := NewCache[string, string]()
	c.Put("key", "value")
	return c
}

func setupNilCache() *Cache[string, string] {
	return &Cache[string, string]{}
}

func TestCache_Get(t *testing.T) {
	type testCase struct {
		name    string
		key     string
		want    []string
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil == no crash",
			key:     "",
			wantErr: true,
		},
		{
			name:    "valid key with return value",
			key:     "key",
			want:    []string{"value"},
			wantErr: false,
		},
		{
			name:    "valid key with no return value",
			key:     "wrong key",
			wantErr: true,
		},
	}

	c := setupCache()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_GetAll(t *testing.T) {
	type testCase struct {
		name    string
		c       *Cache[string, string]
		want    map[string][]string
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil map == no crash",
			c:       setupNilCache(),
			wantErr: true,
		},
		{
			name:    "valid map returned",
			c:       setupCache(),
			want:    map[string][]string{"key": {"value"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_Put(t *testing.T) {
	type args struct {
		key string
		val string
	}
	type testCase struct {
		name    string
		c       *Cache[string, string]
		args    args
		want    []string
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil map still returns value",
			c:       setupNilCache(),
			args:    args{key: "key", val: "value"},
			want:    []string{"value"},
			wantErr: false,
		},
		{
			name:    "valid map with non existing key returned",
			c:       setupCache(),
			args:    args{key: "key2", val: "value"},
			want:    []string{"value"},
			wantErr: false,
		},
		{
			name:    "valid map with existing key returned",
			c:       setupCache(),
			args:    args{key: "key", val: "value2"},
			want:    []string{"value", "value2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Put(tt.args.key, tt.args.val)

			got, err := tt.c.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
