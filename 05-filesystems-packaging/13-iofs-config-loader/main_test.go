package main

import (
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestLoadConfigs(t *testing.T) {
	type args struct {
		fsys fs.FS
		root string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "",
			args: args{
				fsys: fstest.MapFS{
					"config/db.conf":      &fstest.MapFile{Data: []byte("db_host=localhost")},
					"config/api.conf":     &fstest.MapFile{Data: []byte("api_key=secret")},
					"config/readme.txt":   &fstest.MapFile{Data: []byte("ignore me")},
					"config/sub/app.conf": &fstest.MapFile{Data: []byte("app_name=kata")},
				},
				root: "config",
			},
			want: map[string][]byte{
				"config/db.conf":      []byte("db_host=localhost"),
				"config/api.conf":     []byte("api_key=secret"),
				"config/sub/app.conf": []byte("app_name=kata"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfigs(tt.args.fsys, tt.args.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfigs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
