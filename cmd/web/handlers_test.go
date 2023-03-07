package main

import (
	"net/http"
	"snippetbox/internal/assert"
	"testing"
)

func TestPing(t *testing.T) {
	// create a new instance of our application struct. For now, this just contains
	// a couple of mock logger (which discard anything written to them)
	app := newTestApplication(t)

	// we then use the httptest.NewTLSSserver() function to create a new test server
	// passing in the value returned by our app.routes() method as the handler for the server
	// This starts up a HTTPS server which listens on a randomly-chosen port of your local machine
	// for the duration of the test.
	// Notice that we defer a call to ts.Close() so that the server is shutdown when the test finishes
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}
