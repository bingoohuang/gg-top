package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"time"

	"github.com/bingoohuang/gg/pkg/codec"

	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/sigx"
)

//go:embed spline-chart
var splitChart embed.FS

var (
	pInterval = fla9.Duration("interval", 5*time.Minute, "collect interval, eg. 5m")
	pInit     = fla9.Bool("init", false, "create initial ctl")
	pVersion  = fla9.Bool("version", false, "show version and exit")
	pPids     = fla9.String("pids", "", "pids, like 10,12")
)

func main() {
	fla9.Parse()
	ctl.Config{Initing: *pInit, PrintVersion: *pVersion}.ProcessInit()
	ctx, _ := sigx.RegisterSignals(context.Background())

	go collect(ctx, *pInterval, *pPids)

	serverRoot, err := fs.Sub(splitChart, "spline-chart")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		mux.Handle("/", http.FileServer(http.FS(serverRoot)))
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

func collect(ctx context.Context, interval time.Duration, pids string) {
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
	file       *os.File
)

func doCollect(interval time.Duration, pids string) {
	if pids == "" {
		pids = strconv.Itoa(os.Getpid())
	}
	out, err := exec.Command("sh", "-c", "top -bn1 -p "+pids).Output()
	if err != nil {
		log.Printf("exec failed, error:%v", err)
		return
	}

	t := time.Now().Truncate(interval).Format(`2006-01-02T15:04:05`)
	fields, result := ExtractTop(t, string(out))

	if !reflect.DeepEqual(lastFields, fields) {
		fmt.Printf("// %s\n", codec.Json(fields))
		lastFields = fields
		if file != nil {
			file.Close()
		}
		file, _ = os.Open("gg-top" + t + ".data")
		ioutil.WriteFile("gg-top"+t+".json", codec.Json(result), os.ModePerm)
	}

	if file != nil {
		file.WriteString(result + "\n")
	}

	fmt.Printf("%s\n", result)
}
