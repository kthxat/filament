package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kthxat/filament/config"
	"github.com/kthxat/filament/frontend"
	"github.com/davecgh/go-spew/spew"
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

	fmt.Println(appName)
	if len(appVersion) > 0 {
		fmt.Printf("\tVersion %\n", appVersion)
	}
	fmt.Printf("\t\u00a9 %s %s\n", yearStr, appAuthor)
	fmt.Println()
}

func main() {
	printHeader()
	config.ReadConfig(appID)
	spew.Dump(config.GetConfig())

	server := frontend.NewFrontendServer(config.GetConfig().HTTP)
	go server.ListenAndServe()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	<-done

	server.Close()
}
