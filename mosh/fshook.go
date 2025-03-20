// Copyright 2019 Setin Sergei
// Licensed under the Apache License, Version 2.0 (the "License")

package mosh

import (
	"net/http"
	"os"
	"strings"
)

// fsHook wraps an HTTP handler to provide custom 404 handling
type fsHook struct {
	handler  http.Handler
	basePath string
}

// serve404 serves the 404.html page if available
func serve404(basePath string, w http.ResponseWriter) {
	if data, err := os.ReadFile(basePath + "404.html"); err == nil {
		_, _ = w.Write(data)
	}
}

// hookedResponseWriter intercepts HTTP responses to serve custom 404 pages
type hookedResponseWriter struct {
	http.ResponseWriter
	basePath string
	ignore   bool
}

func (hrw *hookedResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		hrw.Header().Set("Content-Type", "text/html")
		hrw.ResponseWriter.WriteHeader(status)
		serve404(hrw.basePath, hrw)
		hrw.ignore = true
		return
	}
	hrw.ResponseWriter.WriteHeader(status)
}

func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
	if hrw.ignore {
		return len(p), nil
	}
	return hrw.ResponseWriter.Write(p)
}

func (fs *fsHook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		serve404(fs.basePath, w)
		return
	}
	fs.handler.ServeHTTP(&hookedResponseWriter{ResponseWriter: w, basePath: fs.basePath}, r)
}

// NewFsHook initializes a new file server with custom 404 handling
func NewFsHook(basePath string) *fsHook {
	return &fsHook{
		handler:  http.FileServer(http.Dir(basePath)),
		basePath: basePath,
	}
}
