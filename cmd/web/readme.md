package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// Define an application struct to hold the application-wide dependencies for the web application
// For now we'll only inlcude fields for the two custom loggers, but we'll add more to it
// as the build progresses
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {

	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the flag
	// will be stored in the addr variable at runtime.
	addr := flag.String("addr", ":4000", "HTTP network address")

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

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	mux := http.NewServeMux()

	// Create a file server which serves files out of the "./ui/static directory."
	// Note that the path given to the http.Dir function is relative to the project
	// directory root
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching parts, we strip the "/static"
	// prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Register the other application routes as normal

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)

	// Initialize a new http.Server struct. We set the Addr and Handler fields so
	// that the server uses the same network address and routes as before, and set
	// the ErrorLog field so that the server now uses the custom errlorLog in
	// the event of any problems
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	// // Create a log file: (not recommended)
	// f, errOp := os.OpenFile("/home/xt03et/go/workspace/my.log", os.O_RDWR|os.O_CREATE, 0666)
	// if errOp != nil {
	// 	log.Fatal(errOp)
	// }
	// defer f.Close()

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself
	// So we need to dereference the pointer (i.e. prefix it with the * symbol) before using it.
	// Note that we're using the log.Printf() function to interpolate the address with the log message.
	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}


### HANDLERS.GO
package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox/internal/models" // New import
	"strconv"
)

// you can think of handler as being a controller
// Generally speaking, they're responsible for carrying out your application logic and writing response headers and bodies.

// Change the signture of the home handler so it is defined as a method against *application
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w) // Use the notFound() helper
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// use the new render helper
	app.render(w, http.StatusOK, "home.tmpl", &templateData{Snippets: snippets})

	// THE CODE IS MADE OBSOLETE BECAUSE OF THE RENDER HELPER
	// files := []string{
	// 	"./ui/html/base.tmpl",
	// 	"./ui/html/partials/nav.tmpl",
	// 	"./ui/html/pages/home.tmpl",
	// }

	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// //Create an instance of a templateData struct holding the slice of snippets
	// data := &templateData{
	// 	Snippets: snippets,
	// }

	// // // Use the ExecuteTemplate() method to write the content of the "base" template as the response body
	// err = ts.ExecuteTemplate(w, "base", data)
	// // also update the code here to use the error logger from the application struct
	// if err != nil {
	// 	app.errorLog.Print(err.Error())
	// }

}

// Change the signature of the snippetView handler so it is defined as a method
// against *application
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w) // Use the notFound() helper.
		return
	}

	// Use the SnippetModel object's Get Method to retrieve the data for a specific record based on its ID.
	// If no matching record is found, return 404 Not Found response
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// use the new render helper
	app.render(w, http.StatusOK, "view.tmpl", &templateData{Snippet: snippet})

	// THE CODE IS MADE OBSOLETE BECAUSE OF THE RENDER HELPER
	// // Initialize a slice containing the paths to the two files. It's important to note that
	// // the file containing our base template must be the *first* file in the slice
	// files := []string{
	// 	"./ui/html/base.tmpl",
	// 	"./ui/html/partials/nav.tmpl",
	// 	"./ui/html/pages/view.tmpl",
	// }

	// // Parse the template files
	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// // Create an instance of a templateData struct holding the snippe data
	// data := &templateData{
	// 	Snippet: snippet,
	// }

	// // And then execute them. Notice how we are passing in the snippet data (a models.Snippet struct)as the final parameter?
	// // err = ts.ExecuteTemplate(w, "base", snippet)

	// // Pass in the templateData struct when executing the template
	// err = ts.ExecuteTemplate(w, "base", data)
	// if err != nil {
	// 	app.serverError(w, err)
	// }

} // end of snippetView

// Change the signature of the snippetCreate handler so it is defined as a method
// against *application
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed) // use the clientError
		return
	}
	// Create some variables holding dummy date. We'll remove these later on during the build
	title := "0 snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	// Pass the date to the SnippetModel.Insert9) method receiving the ID of the new record back.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)

} // end of snippetCreate
