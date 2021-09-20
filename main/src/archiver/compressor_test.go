package archiver

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"hash"
	"os"
	"reflect"
	"testing"
)

func TestCompressor_LoadPaths(t *testing.T) {
	type fields struct {
		incl    []PathInfo
		excl    []PathInfo
		relRoot string
		fw      *os.File
		gw      *gzip.Writer
		tw      *tar.Writer
		hasher  hash.Hash
	}
	type args struct {
		paths  []string
		isIncl bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "1", fields: fields{
			incl:    nil,
			excl:    nil,
			relRoot: "",
			fw:      nil,
			gw:      nil,
			tw:      nil,
			hasher:  nil,
		}, args: args{
			paths:  []string{"/opt", "/proc", "/opt", "/opt/oss", "/home", "/usr"},
			isIncl: false,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compressor{
				incl:    tt.fields.incl,
				excl:    tt.fields.excl,
				relRoot: tt.fields.relRoot,
				fw:      tt.fields.fw,
				gw:      tt.fields.gw,
				tw:      tt.fields.tw,
				hasher:  tt.fields.hasher,
			}
			c.LoadPaths(tt.args.paths, tt.args.isIncl)
			fmt.Println(c.excl)
		})
	}
}

func Test_removeChildPath(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name         string
		args         args
		wantParPaths []string
	}{
		{name: "1", args: args{paths: []string{"/opt", "/opt/oss"}}, wantParPaths: []string{"/opt"}},
		{name: "2", args: args{paths: []string{"/", "/opt", "/opt/oss"}}, wantParPaths: []string{"/"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotParPaths := removeChildPath(tt.args.paths); !reflect.DeepEqual(gotParPaths, tt.wantParPaths) {
				t.Errorf("removeChildPath() = %v, want %v", gotParPaths, tt.wantParPaths)
			}
		})
	}
}

func TestCompressor_addAllPredecessors(t *testing.T) {
	type fields struct {
		incl    []PathInfo
		excl    []PathInfo
		relRoot string
		fw      *os.File
		gw      *gzip.Writer
		tw      *tar.Writer
		hasher  hash.Hash
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "1", fields: fields{
			incl: []PathInfo{{
				path:  "/opt/oss/manager",
				isRec: false,
			},
				{
					path:  "/backup",
					isRec: false,
				},
			},
			excl:    nil,
			relRoot: "",
			fw:      nil,
			gw:      nil,
			tw:      nil,
			hasher:  nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compressor{
				incl:    tt.fields.incl,
				excl:    tt.fields.excl,
				relRoot: tt.fields.relRoot,
				fw:      tt.fields.fw,
				gw:      tt.fields.gw,
				tw:      tt.fields.tw,
				hasher:  tt.fields.hasher,
			}
			c.AddAllPredecessors()
			fmt.Println(c.incl)
		})
	}
}
