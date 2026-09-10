// Code generated for package schema by go-bindata DO NOT EDIT. (@generated)
// sources:
// mutation.graphql
// query.graphql
// schema.graphql
// type/job.graphql
package schema

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _mutationGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2a\xa9\x2c\x48\x55\xf0\x2d\x2d\x49\x2c\xc9\xcc\xcf\x53\xa8\xe6\xe5\x52\x50\x50\x56\x70\x2e\x4a\x4d\x2c\x49\x55\xf0\x4b\x2d\x57\xf0\xca\x4f\x02\x89\xb9\xe6\x15\x96\xa6\x96\xa6\x6a\x94\x24\x16\x67\x5b\x29\x04\x97\x14\x65\xe6\xa5\x2b\x6a\x5a\x81\xa4\x15\x79\xb9\x6a\x01\x01\x00\x00\xff\xff\xe2\x1b\x2e\x85\x46\x00\x00\x00")

func mutationGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_mutationGraphql,
		"mutation.graphql",
	)
}

func mutationGraphql() (*asset, error) {
	bytes, err := mutationGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "mutation.graphql", size: 70, mode: os.FileMode(438), modTime: time.Unix(1749630516, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _queryGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x44\xcb\xb1\xca\xc2\x30\x14\xc5\xf1\xbd\xd0\x77\x38\xe1\x5b\x3e\x5f\x21\xb3\x08\x16\x17\xc9\x28\x0e\x49\xbd\x34\x91\x98\x48\xef\x0d\x12\xc4\x77\x97\xd4\x8a\xdb\xe1\xf0\xff\x49\xbd\x13\x8e\x85\xe6\x8a\x67\xdf\x01\x7f\x38\x04\x16\xd8\x18\x71\xcd\x8e\x11\x12\xc4\x13\xb8\xb2\xd0\xad\x05\x43\x76\xac\x71\x1a\xb2\x53\x67\xd5\x77\x1f\xb3\x23\x19\x3d\x2c\x38\xa4\x29\x52\x93\x70\x15\xfb\xed\x0a\xfe\xc3\x45\xc3\xc8\x1c\xd2\xa4\x36\xba\x3d\x5f\x68\x7c\x7e\x2c\x39\x8b\x95\xc2\x18\x73\x49\xc2\x2b\x33\xcb\xa7\x7f\x53\xf5\xdd\xeb\x1d\x00\x00\xff\xff\x45\x42\x86\x23\xb1\x00\x00\x00")

func queryGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_queryGraphql,
		"query.graphql",
	)
}

func queryGraphql() (*asset, error) {
	bytes, err := queryGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "query.graphql", size: 177, mode: os.FileMode(438), modTime: time.Unix(1749630495, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2a\x4e\xce\x48\xcd\x4d\x54\xa8\xe6\xe5\x52\x50\x50\x50\x28\x2c\x4d\x2d\xaa\xb4\x52\x08\x04\x51\x10\x91\xdc\xd2\x92\xc4\x92\xcc\xfc\x3c\x2b\x05\x5f\x28\x8b\x97\xab\x16\x10\x00\x00\xff\xff\x3d\xf9\x99\x58\x35\x00\x00\x00")

func schemaGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_schemaGraphql,
		"schema.graphql",
	)
}

func schemaGraphql() (*asset, error) {
	bytes, err := schemaGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema.graphql", size: 53, mode: os.FileMode(438), modTime: time.Unix(1749539727, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _typeJobGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x8c\x41\x0a\x83\x40\x10\x04\xcf\x0e\xcc\x1f\xda\x6f\xf8\x83\xe4\xea\x0b\x36\xd9\x89\x2c\xd1\x71\x71\xda\x43\x08\xfe\x3d\x24\x20\x9a\x63\x15\x45\xf1\x55\x0d\xd7\xf9\x86\xb7\x4a\x53\x72\x87\x9e\x4b\xf1\xa1\x55\x69\x98\xe2\x79\xe6\x60\xe2\x1a\x67\x93\x48\x9b\x2a\xa3\xc3\xc5\xd9\xaa\x6c\x2a\x2a\xfb\xb1\xff\xe5\xdf\x2f\x50\xcd\x73\xf1\x61\xef\x80\x65\x75\xff\x13\x8f\x54\x46\xcb\x07\xdf\xe7\xa9\x8e\xc6\x43\x6d\x9f\x00\x00\x00\xff\xff\xa4\xd0\xf3\x9a\xaa\x00\x00\x00")

func typeJobGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_typeJobGraphql,
		"type/job.graphql",
	)
}

func typeJobGraphql() (*asset, error) {
	bytes, err := typeJobGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "type/job.graphql", size: 170, mode: os.FileMode(438), modTime: time.Unix(1749625917, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"mutation.graphql": mutationGraphql,
	"query.graphql":    queryGraphql,
	"schema.graphql":   schemaGraphql,
	"type/job.graphql": typeJobGraphql,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"mutation.graphql": &bintree{mutationGraphql, map[string]*bintree{}},
	"query.graphql":    &bintree{queryGraphql, map[string]*bintree{}},
	"schema.graphql":   &bintree{schemaGraphql, map[string]*bintree{}},
	"type": &bintree{nil, map[string]*bintree{
		"job.graphql": &bintree{typeJobGraphql, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
