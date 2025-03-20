// Copyright 2019 Setin Sergei
// Licensed under the Apache License, Version 2.0 (the "License")

package mosh

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"moshcast/log"
	"moshcast/pool"

	"github.com/gorilla/mux"
)

const (
	cServerName = "MoshCast"
	cVersion    = "1.0.0"
)

type Server struct {
	serverName string
	version    string

	Options Options

	mux            sync.Mutex
	Started        int32
	StartedTime    time.Time
	ListenersCount int32
	SourcesCount   int32

	srv         *http.Server
	poolManager PoolManager
	logger      Logger
	Database    *sql.DB
	SQL         SQL
}

// Init - Load params from config.yaml
func NewServer() (*Server, error) {
	srv := &Server{
		serverName:  cServerName,
		version:     cVersion,
		poolManager: pool.NewPoolManager(),
	}

	err := srv.Options.Load()
	if err != nil {
		return nil, err
	}
	srv.logger, err = log.NewLogger(srv.Options.Logging.LogLevel, srv.Options.Paths.Log)
	if err != nil {
		return nil, err
	}

	db := SQL{
		DBFile: srv.Options.Auth.DBFile,
		logger: srv.logger,
	}
	sqlDB, err := db.Init()
	if err != nil {
		return nil, err
	}
	db.CreateTable(sqlDB)
	srv.Database = sqlDB
	srv.SQL = db

	err = srv.initMounts()
	if err != nil {
		return nil, err
	}

	srv.logger.Log("%s %s", srv.serverName, srv.version)

	srv.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(srv.Options.Socket.Port),
		Handler: srv.configureRouter(),
	}

	return srv, nil
}

func (i *Server) configureRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(i.middlewareHandler)
	r.StrictSlash(true)

	for _, mnt := range i.Options.Mounts {
		r.HandleFunc("/"+mnt.Name, mnt.write).Methods("SOURCE", "PUT")
		r.HandleFunc("/"+mnt.Name, mnt.read).Methods("GET")
		r.Path("/admin/metadata").Queries("mode", "updinfo", "mount", "/"+mnt.Name).HandlerFunc(mnt.meta).Methods("GET")
	}

	r.HandleFunc("/meta.json", i.metaHandler).Methods("GET")
	r.HandleFunc("/info", i.infoHandler).Methods("GET")
	r.HandleFunc("/info.json", i.jsonHandler).Methods("GET")
	r.PathPrefix("/").Handler(NewFsHook(i.Options.Paths.Web))

	return r
}

func (i *Server) initMounts() error {
	var err error
	for idx := range i.Options.Mounts {
		err = i.Options.Mounts[idx].Init(i, i.logger, i.poolManager)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Server) incListeners() {
	atomic.AddInt32(&i.ListenersCount, 1)
}

func (i *Server) decListeners() {
	if atomic.LoadInt32(&i.ListenersCount) > 0 {
		atomic.AddInt32(&i.ListenersCount, -1)
	}
}

func (i *Server) checkListeners() bool {
	clientsLimit := atomic.LoadInt32(&i.Options.Limits.Clients)
	return atomic.LoadInt32(&i.ListenersCount) <= clientsLimit
}

func (i *Server) incSources() {
	atomic.AddInt32(&i.SourcesCount, 1)
}

func (i *Server) decSources() {
	if atomic.LoadInt32(&i.SourcesCount) > 0 {
		atomic.AddInt32(&i.SourcesCount, -1)
	}
}

func (i *Server) checkSources() bool {
	sourcesLimit := atomic.LoadInt32(&i.Options.Limits.Sources)
	return atomic.LoadInt32(&i.SourcesCount) <= sourcesLimit
}

// Close - finish
func (i *Server) Close() {
	if err := i.srv.Shutdown(context.Background()); err != nil {
		i.logger.Error(err.Error())
		i.logger.Log("Error: %s\n", err.Error())
	} else {
		i.logger.Log("Stopped")
	}

	for idx := range i.Options.Mounts {
		i.Options.Mounts[idx].Close()
	}

	i.logger.Close()
}

func (i *Server) writeAccessLog(host string, startTime time.Time, request string, bytesSend int, refer, userAgent string, seconds int) {
	i.logger.Access("%s - - [%s] \"%s\" %s %d \"%s\" \"%s\" %d\r\n", host, startTime.Format(time.RFC1123Z), request, "200", bytesSend, refer, userAgent, seconds)
}

func (i *Server) middlewareHandler(next http.Handler) http.Handler {
	// Ensure that the server will restart after conditions is true (e.g. Docker restart?)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errStr := fmt.Sprintf("%v", err)
				if strings.Contains(errStr, "index out of range") {
					i.logger.Error("%s", err)
					i.logger.Log("%s", err)
					os.Exit(1)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (i *Server) getHost(addr string) string {
	idx := strings.Index(addr, ":")
	if idx == -1 {
		return addr
	}
	return addr[:idx]
}

/*Start - start listening port ...*/
func (i *Server) Start() {
	if atomic.LoadInt32(&i.Started) == 1 {
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		i.mux.Lock()
		i.StartedTime = time.Now()
		i.mux.Unlock()
		atomic.StoreInt32(&i.Started, 1)
		i.logger.Log("Started on %s", i.srv.Addr)

		if err := i.srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-stop
	atomic.StoreInt32(&i.Started, 0)
}
