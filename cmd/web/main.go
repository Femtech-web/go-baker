// import my packages
package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/femtech-web/baker/internal/models"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// declare my app struct
type application struct {
	errorLogger    *log.Logger
	infoLogger     *log.Logger
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debug          bool
	templateCache  map[string]*template.Template
	users          *models.UserModel
}

// declare my main func then in it:
func main() {
	// -- get any cmd flags and parse them
	dsn := flag.String("dsn", "web:web@/baker?parseTime=true&multiStatements=true", "MYSQL datasource name")
	addr := flag.String("addr", ":4000", "port number")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	// -- initialize my error and info loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// -- call my openDB func and defer a "close" func
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	// -- initialize my templatecache
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// -- initialize all dependencies for my app struct
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// -- initialize my app struct with dependencies
	app := &application{
		errorLogger:    errorLog,
		infoLogger:     infoLog,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		debug:          *debug,
		templateCache:  templateCache,
		users:          &models.UserModel{DB: db},
	}

	// todo: initialize dependencies for my server struct

	// -- initialize server struct with dependencies
	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// -- log server details and listen on server
	infoLog.Printf("server running on port: %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// declare my openDB func to open mysql database
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
