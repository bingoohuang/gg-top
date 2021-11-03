package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bingoohuang/gg/pkg/filex"
	"github.com/bingoohuang/gg/pkg/handy"
	"github.com/bingoohuang/gg/pkg/httpp"
	"github.com/bingoohuang/gg/pkg/ss"

	"github.com/bingoohuang/gg/pkg/codec"

	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/sigx"
)

//go:embed spline-chart
var splitChart embed.FS

var (
	pInterval = fla9.Duration("interval", 1*time.Minute, "collect interval, eg. 5m")
	pInit     = fla9.Bool("init", false, "create initial ctl")
	pFile     = fla9.String("file", "", "data file")
	pVersion  = fla9.Bool("version", false, "show version and exit")
	pPids     = fla9.String("pids", "", "pids, like 10,12")
)

func pFileExists() bool {
	if *pFile == "" {
		return false
	}

	f := strings.TrimSuffix(filepath.Base(*pFile), filepath.Ext(*pFile))
	return filex.Exists(f+".json") && filex.Exists(f+".data")
}

func main() {
	fla9.Parse()
	ctl.Config{Initing: *pInit, PrintVersion: *pVersion}.ProcessInit()
	ctx, _ := sigx.RegisterSignals(context.Background())

	if !pFileExists() {
		go collect(ctx, *pInterval, ss.Split(*pPids, ss.WithSeps(",")))
	}

	serverRoot, err := fs.Sub(splitChart, "spline-chart")
	if err != nil {
		log.Fatal(err)
	}
	handler := http.FileServer(http.FS(serverRoot))

	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		mux.Handle("/", ggHandle(handler))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("exiting...")

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func ggHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/js/data.js" {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			httpp.NoCacheHeaders(w, r)

			if dataFileExists() {
				w.Write(createData())
				return
			}

			if pFileExists() {
				data, err := readDataFile(*pFile)
				if err == nil {
					w.Write(data)
					return
				}

				log.Printf("failed to read file %s, error: %v", *pFile, err)
			}
		}

		h.ServeHTTP(w, r)
	})
}

func collect(ctx context.Context, interval time.Duration, pids []string) {
	if interval == 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	doCollect(interval, pids)

	for {
		select {
		case <-ticker.C:
			doCollect(interval, pids)
		case <-ctx.Done():
			return
		}
	}
}

var (
	lastFields []string
	file       string
	fileLock   sync.Mutex
)

func dataFileExists() bool {
	defer handy.LockUnlock(&fileLock)()

	return filex.Exists(file)
}

func createData() []byte {
	defer handy.LockUnlock(&fileLock)()

	var b bytes.Buffer
	b.WriteString(`const headers = `)
	b.Write(codec.Json(lastFields))
	b.WriteString("\nconst data = [")
	data, _ := ioutil.ReadFile(file)
	b.Write(bytes.TrimRight(data, "\n,"))
	b.WriteString("]\n")

	return b.Bytes()
}

func readDataFile(filename string) ([]byte, error) {
	f := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	header, err := ioutil.ReadFile(f + ".json")
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(f + ".data")
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.WriteString(`const headers = `)
	b.Write(header)
	b.WriteString("\nconst data = [")
	b.Write(bytes.TrimRight(data, "\n,"))
	b.WriteString("]\n")

	return b.Bytes(), nil
}

func doCollect(interval time.Duration, pids []string) {
	if len(pids) == 0 {
		pids = []string{strconv.Itoa(os.Getpid())}
	}
	out, err := exec.Command("sh", "-c", topCmd(pids)).Output()
	if err != nil {
		log.Printf("exec failed, error:%v", err)
		return
	}

	t := time.Now().Truncate(interval).Format(`2006-01-02T15:04:05`)
	fields, result := ExtractTop(t, string(out))

	defer handy.LockUnlock(&fileLock)()

	if !reflect.DeepEqual(lastFields, fields) {
		fmt.Printf("%s\n", codec.Json(fields))
		lastFields = fields
		tt := ss.Strip(t, func(r rune) bool { return !unicode.IsDigit(r) })
		file = "gg-top-" + tt + ".data"
		ioutil.WriteFile("gg-top-"+tt+".json", codec.Json(fields), os.ModePerm)
	}

	filex.Append(file, []byte(result+",\n"))
	fmt.Printf("%s\n", result)
}
