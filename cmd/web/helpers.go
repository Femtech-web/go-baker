package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLogger.Println(trace)
	app.errorLogger.Output(2, trace)

	if app.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: app.isAuthenticated(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		CSRFToken:       nosurf.Token(r),
	}
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) decodeForm(r *http.Request, t string, dst any) error {
	var err error

	if t == "multipart" {
		err = r.ParseMultipartForm(32 << 20)
	} else {
		err = r.ParseForm()
	}

	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)

	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *application) currentUser(r *http.Request) int {
	return app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
}

func (app *application) csvMap(header, columns []string) (map[string]int, []string) {
	csvToDBMap := make(map[string]int)
	features := make([]string, 0)

	for i, colName := range header {
		for _, dbColumn := range columns {
			if strings.Contains(dbColumn, "feature") {
				feature := strings.Split(dbColumn, "_")[1]
				n := strings.Split(dbColumn, "_")[0]

				if colName == feature {
					csvToDBMap[n] = i
					features = append(features, feature)
				}
			} else {
				if colName == dbColumn {
					csvToDBMap[colName] = i
				}
			}
		}
	}

	return csvToDBMap, features
}
