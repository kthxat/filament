package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kthxat/filament/config"
	"github.com/kthxat/filament/frontend"
)

const (
	appID          = "filament"
	appName        = "Filament"
	appAuthor      = `Carl Kittelberger`
	appDescription = "HTTP-to-FTP server."
)

var (
	appDevelopmentStartTime = mustParseTime(time.Parse(time.RFC1123, "Fri, 05 Apr 2019 00:00:00 CET"))
	appBuildTime            = time.Now()
	appVersion              = ""
)

func mustParseTime(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func printHeader() {
	var yearStr string
	if appBuildTime.Year() > appDevelopmentStartTime.Year() {
		yearStr = fmt.Sprintf("%d\u2013%d", appDevelopmentStartTime.Year(), appBuildTime.Year())
	} else {
		yearStr = fmt.Sprintf("%d", appBuildTime.Year())
	}

	fmt.Fprintln(os.Stderr, appName)
	if len(appVersion) > 0 {
		fmt.Fprintf(os.Stderr, "\tVersion %s\n", appVersion)
	}
	fmt.Fprintf(os.Stderr, "\t\u00a9 %s %s\n", yearStr, appAuthor)
	fmt.Fprintln(os.Stderr)
}

func main() {
	printHeader()
	config.ReadConfig(appID)
	spew.Dump(config.GetConfig())

	server := frontend.NewFrontendServer(config.GetConfig().HTTP)
	go func() {
		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			return
		} else if err != nil {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	<-done

	if err := server.Close(); err != nil {
		log.Printf("Failed to cleanly shut down server: %s",
			err.Error())
	}
}
