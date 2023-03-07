package main

import (
	"net/http"
	"snippetbox/ui"

	"github.com/julienschmidt/httprouter" // New import
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Initialize the router.
	router := httprouter.New()

	// Create a handler function which wraps our notFound() helper, and then assign it as the
	// custom handler for 404 Not Found responses. You can also set a custom handler for 405
	// Method Not Allowed responses by setting router.MethodNotAllowed in the same way too
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Take the ui.Files embedded filesystem and convert it to a http.FS type so that it
	// satisfies the http.FileSystem interface. We then pass that to the http.FileServer()
	// function to create the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))

	// Our static files are contained in the "static" folder of the ui.Files embedded filesystem
	// So, for example, our CSS stylesheet is located at "static/css/main.css". This means that
	// we now no longer need to strip the prefix from the request URL -- any requests that start
	// with /static can just be passed directly to the file server and the corresponding static
	// file will be served (so long as it exists)
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// Add n new GET /ping route
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// Use the nosurf middleware on all our 'dynamic' routes.
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// Update the routes to use the new dynamic middleware chain followed by the appropriate
	// handler function. Note that because the alice ThenFunc() method returns a http.Handler
	// (rather than a http.HandlerFunc) we also need to switch to registering the route using the
	// router.Handler() method.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	// Add the five new routes, all of which use our 'dynamic' middleware chain
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// Because the 'protected' middleware chain appends to the 'dynamic' chain
	// the noSurf middleware will also be sued on the three routes below too
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// Create the middleware chain as normal.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Wrap the router with the middleware and return it as normal.
	return standard.Then(router)
}
