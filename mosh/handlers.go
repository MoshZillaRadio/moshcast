// Copyright 2019 Setin Sergei
// Licensed under the Apache License, Version 2.0 (the "License")

package mosh

import (
	"html/template"
	"net/http"
	"os"
)

// internalHandler serves the internal server error page
func (s *Server) internalHandler(w http.ResponseWriter) {
	if content, err := os.ReadFile(s.Options.Paths.Web + "500.html"); err == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(content)
	} else {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// infoHandler renders the info page
func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, "templates/info.gohtml", nil)
}

// jsonHandler renders the JSON response page
func (s *Server) jsonHandler(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, "templates/json.gohtml", nil)
}

// metaHandler renders the meta page
func (s *Server) metaHandler(w http.ResponseWriter, r *http.Request) {
	dateParam := r.URL.Query().Get("date")
	if len(dateParam) == 0 {
		dateParam = "2025-03-03"
	}
	data := s.SQL.SelectMetaData(s.Database, dateParam)
	s.renderPage(w, "templates/meta.gohtml", data)
}

func (s *Server) renderPage(w http.ResponseWriter, tplName string, data interface{}) {
	if data == nil {
		data = &s
	}

	t, err := template.ParseFiles(tplName)
	if err != nil {
		s.logger.Error("Failed to parse template: " + err.Error())
		s.internalHandler(w)
		return
	}
	if err := t.Execute(w, data); err != nil {
		s.logger.Error("Failed to execute template: " + err.Error())
		s.internalHandler(w)
	}
}
