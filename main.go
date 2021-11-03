package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
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
	pFile     = fla9.String("file", "", "data file, with :generate to create a zip html file and exit")
	pVersion  = fla9.Bool("version", false, "show version and exit")
	pPids     = fla9.String("pids", "", "pids, like 10,12")

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

func main() {
	if pFileGenerate {
		generateReportTarGz()
		return
	}

	ctx, _ := sigx.RegisterSignals(context.Background())

	if !pFileExists {
		go collectLoop(ctx, *pInterval, ss.Split(*pPids, ss.WithSeps(",")))
	}

	handler := http.FileServer(http.FS(serverRoot))

	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		mux.Handle("/", ggHandle(handler))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	var replaced []byte
	if *pInterval > 0 {
		replaced = []byte(fmt.Sprintf(`<meta http-equiv="refresh" content="%d" >`, *pInterval/time.Second))
	}
	data, hash, contentType, _ := emb.Asset(serverRoot, "index.html", false)
	data = bytes.Replace(data, []byte(`<meta http-equiv="refresh" content="600" >`), replaced, 1)

	httpp.NoCacheHeaders(w, r)
	w.Header().Set("Content-Type", contentType)
	w.Header().Add("ETag", hash)
	_, _ = w.Write(data)
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

func collectLoop(ctx context.Context, interval time.Duration, pids []string) {
	if interval == 0 {
		return
	}

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
	file       string
	fileLock   sync.Mutex
)

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

func readDataFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p := bytes.IndexRune(data, '\n')

	var b bytes.Buffer
	b.WriteString(`const headers = `)
	b.Write(data[:p])
	b.WriteString("\nconst data = [")
	b.Write(data[p : len(data)-2])
	b.WriteString("]\n")

	return b.Bytes(), nil
}

func collect(interval time.Duration, pids []string) {
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
		file = "gg-top-" + tt + ".json"
		_, _ = filex.Append(file, append(codec.Json(fields), '\n'))
	}

	_, _ = filex.Append(file, []byte(result+",\n"))
	fmt.Printf("%s\n", result)
}
