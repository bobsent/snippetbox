package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// The serverError helper writes an error message and stack trace to the errorLog
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// the clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad Request"
// when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a convenience
// wrapper around clientError which sends a 404 Not Found response to the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the appropriate template set from the cache based on the page name (like 'home.tmpl')
	// If no exntry exists in the cache with the provided name, then create a new error and call the
	// serverError() helper that we created earlier and return
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// initialize a new buffer
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the http.ResponseWriter.
	// If there's an error, call our serverError() helper and then return
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc.).
	// If the template is written to the buffer without any error, we are safe to go ahead
	// and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(status)

	// write the contents of the buffer to the http.ResponseWriter. Note: this another time
	// where we pass our http.ResponseWriter to a function that takes an io.Writer.
	buf.WriteTo(w)
}

// Create an newTemplateData() helper, which returns a pointer to a templateData
// struct initialized with the current year.
// Note that we're not using the *http.Request parameterhere at the moment, but we'll do later
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

// decodePostForm() helper method.
// The second parameter here, dst, is the target destination that we want
// to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// createSnippetPost handler
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call decode() on our decoder instance, passing the target destination as the first
	// parameter
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method will return
		// an error with the type *form.InvalidDecoderError. We ure errors.As() to check
		// for this and raise a panic rather than returning the error.
		var InvalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &InvalidDecoderError) {
			panic(err)
		}

		// for all other errors, we return them as normal
		return err
	}

	return nil
} // end of decodePostForm

// Return true if the current requesti s from an authenticated user, otherwise return false
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticted, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticted
}
