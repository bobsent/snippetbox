package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	// Notice how the import path for our driver is prefixed with an underscore?
	// This is because our main.go file doesn’t actually use anything in the mysql package.
	// So if we try to import it normally the Go compiler will raise an error.
	// However, we need the driver’s init() function to run so that it can register itself with the database/sql package.
	// The trick to getting around this is to alias the package name to the blank identifier.
	// This is standard practice for most of Go’s SQL drivers.

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql" // New import

	// import the models package that we just created. You need to prefix this with whatever module path you set up
	// back in chapter 02.01 (Project Setup and Creating a Module) so that the import statement looks like this:
	// "{your-module-path)/internal/models". If you can't remember what module path you used, you can find it at the
	// thop of go.mod file: "snippetbox.alexedwards.net/internal/models"
	"snippetbox/internal/models"
)

// Define an application struct to hold the application-wide dependencies for the web application
// For now we'll only inlcude fields for the two custom loggers, but we'll add more to it
// as the build progresses
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface // add a snippetsfield to the application struct. This will allow us to make the Snippetmodel object available to our handlers
	users          models.UserModelInterface
	templateCache  map[string]*template.Template // add a templateCache field
	formDecoder    *form.Decoder                 // add a formDecoder field to hold a pointer to a form.Decoder instance
	sessionManager *scs.SessionManager           // add a new sessionManager field to the application sruct
}

func main() {

	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the flag
	// will be stored in the addr variable at runtime.
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:Pyth0n!sta24@/snippetbox?parseTime=true", "MySQL data source name")

	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr variable
	// You need to call this *before* you use the addr variable otherwise it will always
	// contain the default value of ":4000". If any error are encountered during parsing
	// the application will be terminated.
	flag.Parse()

	// Use log.New() to create a logger for writing information messages. This takes three parameters:
	// 1. the destination to write the logs to (os.Stdout)
	// 2. a string prefix for message (INFO followed by tab)
	// 3. and flags to indicate what additional information to include (local data and time).
	// Note that the flags are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages in the same way, but use stderr as the destination
	// and use the log.Lshortfile flag to include the relewvant filename and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// To keep the main() function tidy, I've put the code for creating a connection pool into a separate
	// openDB() function below. We pass openDB() the DSN from the command-line flag
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Initialize a new template cache
	templateCache, err := newTemplateChache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize a decoder instance
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// Make sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web browser
	// when a HTTP connection is being used (amd won't be sent over an unsecure HTTP connection)
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db}, // initialize a models.SnippetModel instance and add it to the application dependencies
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache, // add templateCache to the dependencies
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initializew a tls.Config struct to hold the non-default TLS settings we want the server to use.
	// In this case, the only thing that we're changing is the curve preferences value, so that only elliptic curver with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(), // call the new app.routes() method to get the servermux containing our routes
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself
	// So we need to dereference the pointer (i.e. prefix it with the * symbol) before using it.
	// Note that we're using the log.Printf() function to interpolate the address with the log message.
	infoLog.Printf("Starting server on %s", *addr)
	// Use the ListenAndServeTLS method to start the HTTPS server. We pass in the paths to the TLS certificate and corresponding
	// private key as the two parameters.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool for a given DSN
func openDB(dsn string) (*sql.DB, error) {
	// The sql.Open() function doesn’t actually create any connections, all it does is initialize the pool for future use.
	// Actual connections to the database are established lazily, as and when needed for the first time.
	// So to verify that everything is set up correctly we need to use the db.Ping() method to create a connection
	// and check for any errors.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
