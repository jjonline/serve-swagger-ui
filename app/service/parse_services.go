package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jjonline/serve-swagger-ui/client"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/jjonline/serve-swagger-ui/define"
	"github.com/jjonline/serve-swagger-ui/utils/memory"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// watch path list
var watchPath = make([]string, 0)

// parseService swagger json parser service
type parseService struct{}

// Data swagger json file Data info
type Data struct {
	Items map[string][]Swagger
	Table map[string]string
}

// Swagger parsed swagger file item
type Swagger struct {
	OpenVersion string // openapi version
	Version     string // swagger doc version
	Title       string // swagger title
	Desc        string // swagger desc
	Name        string // swagger concat name
	Email       string // swagger concat email
	Icon        string // swagger title first letter
	Path        string // json file path detail
	Hash        string // json file path hash
	Key         int    // css icon bg num
}

// define swagger struct
type swagger struct {
	Openapi string `json:"openapi"` // version for openapi up 3.0
	Swagger string `json:"swagger"` // version for below 3.x
	Info    info   `json:"info"`
}
type info struct {
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Title       string  `json:"title"`
	Contact     contact `json:"contact"`
}
type contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ParseWithCache with cache for parse
func (s *parseService) ParseWithCache() (Data, error) {
	path := conf.Config.Swagger.Path
	data, err := memory.GetWithSetter(define.SwaggerCacheKey, func() (interface{}, error) {
		res, err := s.parse(path)
		if err != nil {
			return nil, err
		}
		return res, nil
	}, 0)

	if err != nil {
		return Data{}, err
	}
	if res, ok := data.(Data); ok {
		return res, nil
	}
	return Data{}, err
}

// CleanCache clean parsed cache for reload parse
func (s *parseService) CleanCache() {
	memory.Del(define.SwaggerCacheKey)
}

// FirstDoc get the first doc hash string
func (s *parseService) FirstDoc() (string, error) {
	data, err := s.ParseWithCache()
	if err != nil {
		return "", err
	}

	// get first swagger file
	for hash := range data.Table {
		return hash, nil
	}

	// do not happen
	return "", nil
}

// StartFileWatcher use goroutine watch swagger file changed then clean parsed cache
func (s *parseService) StartFileWatcher() {
	// collect init needed watch dir
	watchPath = s.collectWatchDir(conf.Config.Swagger.Path)

	// define watch func
	var addWatch = func(pathItems []string, watcher *fsnotify.Watcher) {
		for _, path := range pathItems {
			_ = watcher.Add(path)
		}
	}
	var removeWatch = func(pathItems []string, watcher *fsnotify.Watcher) {
		for _, path := range pathItems {
			_ = watcher.Remove(path)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()
	if err != nil {
		panic(err)
	}

	// init watch
	addWatch(watchPath, watcher)

	for {
		select {
		case event, ok := <-watcher.Events:
			if ok {
				client.Logger.Infof("recognize %s changed, auto reread", event.Name)
				// check watch path is changed or not
				newWatchPath := s.collectWatchDir(conf.Config.Swagger.Path)
				if s.isWatchDirChanged(watchPath, newWatchPath) {
					client.Logger.Info("subdirectory was changed, auto reinitialize watcher")
					removeWatch(watchPath, watcher)
					addWatch(newWatchPath, watcher)
					watchPath = newWatchPath // assign new watch path to old
				}

				// when file change, just clear parsed cache, new cache will auto make when a new http request come
				s.CleanCache()
			}
		case <-watcher.Errors:
			// no code
		}
	}
}

// isWatchDirChanged check if watch path dir is change
func (s *parseService) isWatchDirChanged(prev, next []string) bool {
	if (prev == nil) != (next == nil) {
		return true
	}

	if len(prev) != len(next) {
		return true
	}

	// this line can ensure the next[i] never out of index in for...range loop
	next = next[:len(prev)]
	for i, v := range prev {
		if v != next[i] {
			return true
		}
	}

	return false
}

// collectWatchDir collect need watch path list
func (s *parseService) collectWatchDir(parent string) []string {
	var patCollect = make([]string, 0)
	_ = filepath.WalkDir(parent, func(path string, d fs.DirEntry, err error) error {
		// collect Dir path
		if err == nil && d.IsDir() {
			patCollect = append(patCollect, path)
		}
		return nil
	})

	// sort string slice
	sort.Strings(patCollect)

	return patCollect
}

// parse all swagger json files
func (s *parseService) parse(path string) (Data, error) {
	// res map[string][]Swagger
	// no sub-path as default map key
	// sub-path as group map key
	var res = make(map[string][]Swagger, 0)

	// table map[string]string
	// key for json file dir md5 hash
	// value for json file dir detail
	var table = make(map[string]string, 0)

	// ① read dir collect all json file
	path = strings.TrimRight(path, "/")
	dir, err := os.ReadDir(path)
	if err != nil {
		client.Logger.Errorf("open swagger json file path %s occur error: %s", path, err.Error())
		return Data{}, err
	}

	// ② collect all swagger json file
	for _, target := range dir {
		if target.IsDir() {
			// only collect json suffix
			subSwg := s.collectSubDir(path + "/" + target.Name())
			if len(subSwg) > 0 {
				if _, ok := res[target.Name()]; !ok {
					res[target.Name()] = make([]Swagger, 0)
				}
				res[target.Name()] = append(res[target.Name()], subSwg...)
			}
		} else {
			// only collect json suffix
			if strings.HasSuffix(target.Name(), ".json") {
				if result, err1 := s.parseSwagger(path + "/" + target.Name()); err1 == nil {
					if _, ok := res["default"]; !ok {
						res["default"] = make([]Swagger, 0)
					}
					res["default"] = append(res["default"], result)
				}
			}
		}
	}

	// check exist swagger file
	if len(res) <= 0 {
		client.Logger.Error("none swagger json file found")
		return Data{}, fmt.Errorf("none swagger json file found")
	}

	// set hashTable map
	for _, items := range res {
		for _, item := range items {
			table[item.Hash] = item.Path
		}
	}

	return Data{Items: res, Table: table}, nil
}

// collectSubDir list toml sub dir
func (s *parseService) collectSubDir(subPath string) []Swagger {
	res := make([]Swagger, 0)
	_ = filepath.WalkDir(subPath, func(path string, d fs.DirEntry, err error) error {
		// just collect .json suffix file
		if err == nil && !d.IsDir() {
			if strings.HasSuffix(path, ".json") {
				if result, err1 := s.parseSwagger(path); err1 == nil {
					res = append(res, result)
				}
			}
		}
		return nil
	})
	return res
}

// parseSwagger parse swagger json file info
func (s *parseService) parseSwagger(path string) (res Swagger, err error) {
	var swg swagger
	var stream []byte
	stream, err = os.ReadFile(path)
	if err != nil {
		client.Logger.Errorf("open swagger json file %s occur error: %s", path, err.Error())
		return res, err
	}

	err = json.Unmarshal(stream, &swg)
	if err != nil {
		client.Logger.Errorf("parse swagger json file %s occur error: %s", path, err.Error())
		return res, err
	}

	// Openapi
	if swg.Openapi != "" {
		res.OpenVersion = swg.Openapi
	}
	if swg.Swagger != "" {
		res.OpenVersion = swg.Swagger
	}

	// Version
	res.Version = swg.Info.Version
	if swg.Info.Version == "" {
		res.Version = "None Version"
	}

	// Title
	res.Title = swg.Info.Title
	if swg.Info.Title == "" {
		res.Title = "None Title"
	}

	// Desc
	res.Desc = swg.Info.Description
	if swg.Info.Description == "" {
		res.Desc = "None description"
	}

	// Name
	res.Name = swg.Info.Contact.Name
	if swg.Info.Contact.Name == "" {
		res.Name = "None Author"
	}

	// Email
	res.Email = swg.Info.Contact.Email
	if swg.Info.Contact.Email == "" {
		res.Email = "None Email"
	}

	// Icon
	res.Icon = s.upperIcon(res.Title)

	// Hash && Path
	res.Hash = s.md5(path)
	res.Path = path
	res.Key = rand.Intn(7) // random bg Key

	return
}

// upperIcon set one string to upper if possible
func (s *parseService) upperIcon(str string) string {
	res := strings.Split(str, "")
	ss := []byte(res[0])
	if len(ss) == 1 && ss[0] >= 'a' && ss[0] <= 'z' {
		return strings.ToUpper(res[0])
	}
	return res[0]
}

// md5 generate md5
func (s *parseService) md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte("")))
}
