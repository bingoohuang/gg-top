package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bingoohuang/gg/pkg/mapp"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bingoohuang/jj"

	"github.com/bingoohuang/gg/pkg/netx/freeport"

	"github.com/bingoohuang/gg/pkg/emb"

	"github.com/bingoohuang/gg/pkg/filex"
	"github.com/bingoohuang/gg/pkg/handy"
	"github.com/bingoohuang/gg/pkg/httpp"
	"github.com/bingoohuang/gg/pkg/ss"

	"github.com/bingoohuang/gg/pkg/codec"

	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/sigx"
)

var (
	//go:embed spline-chart
	splitChart embed.FS
	serverRoot fs.FS
)

var (
	pInterval = fla9.Duration("interval", 1*time.Minute, "collect interval, eg. 5m")
	pInit     = fla9.Bool("init", false, "create initial ctl")
	pConvert  = fla9.Bool("convert", false, "Convert G/M to K or just drop if")
	pFile     = fla9.String("file", "", "data file, with :generate to create a zip html file and exit")
	pVersion  = fla9.Bool("version", false, "show version and exit")
	pPidWords = fla9.String("p", "", "pids, like 10,12, or command line words")
	pPort     = fla9.Int("port", 1100, "port")

	pFileExists   bool
	pFileGenerate bool
)

func init() {
	var err error

	serverRoot, err = fs.Sub(splitChart, "spline-chart")
	if err != nil {
		log.Fatal(err)
	}

	fla9.Parse()
	ctl.Config{Initing: *pInit, PrintVersion: *pVersion}.ProcessInit()

	if *pFile != "" {
		if pFileGenerate = strings.HasSuffix(*pFile, ":generate"); pFileGenerate {
			*pFile = strings.TrimSuffix(*pFile, ":generate")
		}
	}

	pFileExists = *pFile != "" && filex.Exists(*pFile)

	if pFileExists {
		*pInterval = 0
	}
}

type tgzFilter struct{}

func (t tgzFilter) Support(name string) bool {
	return name == "js/data.js" || name == "index.html"
}

func (t tgzFilter) Filter(name string, data []byte) ([]byte, error) {
	switch name {
	case "js/data.js":
		return readDataFile(*pFile)
	case "index.html":
		return bytes.Replace(data, []byte(`<meta http-equiv="refresh" content="600" >`), nil, 1), nil
	default:
		return data, nil
	}
}

func generateReportTarGz() {
	f := *pFile
	f = strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)) + ".tgz"

	if err := tgzCreate(serverRoot, f, &tgzFilter{}); err != nil {
		log.Fatalf("failed to create tgz %s, error: %v", f, err)
	}

	log.Printf("%s create successfully", f)
}

var numReq = regexp.MustCompile(`^\d+%`)

func isAllNum(ss []string) bool {
	for _, s := range ss {
		if !numReq.MatchString(s) {
			return false
		}
	}

	return len(ss) > 0
}

func main() {
	if pFileGenerate {
		generateReportTarGz()
		return
	}

	ctx, _ := sigx.RegisterSignals(context.Background())

	if !pFileExists {
		pidWords := ss.Split(*pPidWords, ss.WithSeps(","))
		var pids func() []string
		if len(pidWords) == 0 {
			pids = func() []string { return nil }
		} else if isAllNum(pidWords) {
			pids = func() []string { return pidWords }
		} else {
			thisPid := strconv.Itoa(os.Getpid())
			grepWord := ""
			for _, word := range pidWords {
				grepWord += `|grep -w ` + word
			}
			s := `ps -ef|grep -v grep` + grepWord + `|awk '{print $2}'|xargs|sed 's/ /,/g'`
			pids = func() []string { return collectPids(s, thisPid) }
		}
		go collectLoop(ctx, *pInterval, pids)
	}

	handler := http.FileServer(http.FS(serverRoot))

	mux := http.NewServeMux()
	var srv *http.Server

	if *pPort > 0 {
		addr := fmt.Sprintf(":%d", freeport.PortStart(*pPort))
		srv = &http.Server{Addr: addr, Handler: mux}
		log.Printf("Start to listen on %s", addr)

		go func() {
			mux.Handle("/", ggHandle(handler))
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("listen error: %v", err)
			}
		}()
	}

	<-ctx.Done()
	log.Printf("exiting...")

	if srv != nil {
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}
}

func collectPids(s, excludePid string) []string {
	log.Printf(`start to exec shell "%s"`, s)
	out, err := exec.Command("sh", "-c", s).Output()
	if err != nil {
		log.Printf("exec %s failed, error:%v", s, err)
		return nil
	}

	pids := ss.Split(string(out), ss.WithSeps(","))
	result := make([]string, 0, len(pids))
	for _, pid := range pids {
		if pid != excludePid {
			result = append(result, pid)
		}
	}

	return result
}

func ggHandle(h http.Handler) http.Handler {
	defaultHandle := h.ServeHTTP

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/js/data.js":
			handlerData(w, r, defaultHandle)
		case "/", "/index.html":
			handleIndex(w, r)
		default:
			defaultHandle(w, r)
		}
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data, hash, contentType, _ := emb.Asset(serverRoot, "index.html", false)

	var refreshMeta []byte
	if *pInterval > 0 {
		refreshMeta = []byte(fmt.Sprintf(`<meta http-equiv="refresh" content="%d">`, *pInterval/time.Second))
	}
	data = bytes.Replace(data, []byte(`<meta http-equiv="refresh" content="600">`), refreshMeta, 1)

	var header, lsofs []byte
	if pFileExists {
		header, lsofs, _ = readDataFileHeader(*pFile)
	} else {
		header, lsofs = dataFileHeader()
	}

	data = bytes.Replace(data, []byte(`gg-fields:`), header, 1)
	lsofList := mergeLsofs(lsofs)
	data = bytes.Replace(data, []byte(`lsofs:`), []byte(lsofList), 1)

	if f := r.URL.Query().Get("f"); f != "" {
		if ff := ss.Split(f); len(ff) > 0 {
			rr := fmt.Sprintf(`const tagSuffix = %s`, codec.Json(ff))
			data = bytes.Replace(data, []byte(`const tagSuffix = ["RES", "MEM"]`), []byte(rr), 1)
		}
	}

	httpp.NoCacheHeaders(w, r)
	w.Header().Set("Content-Type", contentType)
	w.Header().Add("ETag", hash)
	_, _ = w.Write(data)
}

func mergeLsofs(lsofs []byte) string {
	var lsofList string
	var lsofMap map[string]string
	_ = json.Unmarshal(lsofs, &lsofMap)

	for _, k := range mapp.KeysSorted(lsofMap) {
		lsofList += lsofMap[k] + "\n"
	}

	return lsofList
}

func handlerData(w http.ResponseWriter, r *http.Request, defaultHandle func(http.ResponseWriter, *http.Request)) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	httpp.NoCacheHeaders(w, r)

	if dataFileExists() {
		_, _ = w.Write(createData())
		return
	}

	if pFileExists {
		data, err := readDataFile(*pFile)
		if err == nil {
			_, _ = w.Write(data)
			return
		}

		log.Printf("failed to read file %s, error: %v", *pFile, err)
	}

	defaultHandle(w, r)
}

func collectLoop(ctx context.Context, interval time.Duration, pids func() []string) {
	if interval == 0 {
		return
	}

	time.Sleep(3 * time.Second) // wait some shell to exit
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		collect(interval, pids)

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

var (
	lastFields []string
	lastLsofs  []byte
	file       string
	fileLock   sync.Mutex
)

func dataFileHeader() ([]byte, []byte) {
	defer handy.LockUnlock(&fileLock)()
	return codec.Json(lastFields), lastLsofs
}

func dataFileExists() bool {
	defer handy.LockUnlock(&fileLock)()

	return filex.Exists(file)
}

func createData() []byte {
	defer handy.LockUnlock(&fileLock)()

	data, err := readDataFile(file)
	if err != nil {
		log.Printf("W! failed to read data file %s, error: %v", file, err)
		return nil
	}

	return data
}

func readDataFileHeader(filename string) ([]byte, []byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	r1 := jj.GetBytes(data, "headers")
	r2 := jj.GetBytes(data, "lsof")
	return []byte(r1.Raw), []byte(r2.Raw), nil
}

func readDataFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.WriteString(`const data = `)
	b.Write(data)

	return b.Bytes(), nil
}

func collect(interval time.Duration, pidsFn func() []string) {
	pids := pidsFn()
	if len(pids) == 0 {
		pids = []string{strconv.Itoa(os.Getpid())}
	} else {
		pids = atLeastTwoTimes(pids)
	}

	log.Printf("start to collect top information for pids %s", strings.Join(pids, ","))

	top := topCmd(pids)
	log.Printf(`start to exec shell "%s"`, top)
	out, err := exec.Command("sh", "-c", top).Output()
	if err != nil {
		log.Printf("exec %s failed, error:%v", top, err)
		return
	}

	t := time.Now().Truncate(interval).Format(`2006-01-02T15:04:05`)
	fields, result := ExtractTop(t, string(out), *pConvert)

	defer handy.LockUnlock(&fileLock)()
	var data bytes.Buffer
	backOffset := int64(0)
	if !reflect.DeepEqual(lastFields, fields) {
		log.Printf("%s\n", codec.Json(fields))
		lastFields = fields
		tt := ss.Strip(t, func(r rune) bool { return !unicode.IsDigit(r) })
		file = "gg-top-" + tt + ".json"
		data.Write([]byte(`{"headers":`))
		data.Write(codec.Json(fields))
		data.Write([]byte(",\"lsof\":"))
		lastLsofs = codec.Json(lsof(pids))
		data.Write(lastLsofs)
		data.Write([]byte(",\"data\":[\n"))
		_, _ = filex.Append(file, data.Bytes())
	} else {
		backOffset = -2
	}

	data.Reset()
	if backOffset < 0 {
		data.Write([]byte(",\n"))
	}
	data.Write([]byte(result))
	data.Write([]byte("]}"))

	_, _ = filex.Append(file, data.Bytes(), filex.WithBackOffset(backOffset))
	log.Printf("%s\n", result)
}

func lsof(pids []string) map[string]string {
	m := make(map[string]string)
	for _, pid := range pids {
		lsofCmd := `lsof -p ` + pid
		log.Printf(`start to exec shell "%s"`, lsofCmd)
		out, err := exec.Command("sh", "-c", lsofCmd).Output()
		if err != nil {
			log.Printf("exec %s failed, error:%v", lsofCmd, err)
			continue
		}
		m[pid] = string(out)
	}

	return m
}

var (
	old     []string
	newPids []string
	times   int
)

func atLeastTwoTimes(pids []string) []string {
	sort.Strings(pids)
	if len(old) == 0 {
		old = pids
		return pids
	}

	if reflect.DeepEqual(old, pids) {
		newPids = nil
		times = 0
		return pids
	}

	if !containsSlice(pids, old) {
		old = pids
		newPids = nil
		times = 0
		return pids
	}

	if len(newPids) == 0 {
		newPids = pids
		times = 1
		return old
	}

	if reflect.DeepEqual(newPids, pids) {
		times++
	} else {
		newPids = pids
		times = 1
		return old
	}

	if times >= 2 {
		old = pids
		newPids = nil
		times = 0
		return pids
	}

	return old
}

func containsItem(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// containsSlice tells if slice s contains all elements of s2.
func containsSlice(s1, s2 []string) bool {
	if len(s1) < len(s2) {
		return false
	}

	if len(s1) == 0 {
		return len(s2) == 0
	}

	for _, e := range s2 {
		if !containsItem(s1, e) {
			return false
		}
	}

	return true
}
