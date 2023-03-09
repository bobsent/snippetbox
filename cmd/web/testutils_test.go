package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"snippetbox/internal/models/mocks"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

// create a newTestApplication helper which returns an instance of our
// application struct containing mocked dependencies
func newTestApplication(t *testing.T) *application {
	// create an instance of the template cache
	templateCache, err := newTemplateChache()
	if err != nil {
		t.Fatal(err)
	}

	// and a form decoder
	formDecoder := form.NewDecoder()

	// and a session manager instancr. note that we use the same settings as production
	// except that we *don't* set a Store for the session manager.
	// If no store is set, the SCS package will default to using a transient in-memory store
	// which is ideal for testing purposes.
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

// Define a custom testServer type which embed a httptest.Server instance
type testServer struct {
	*httptest.Server
}

// Create a newTestServer helper which initializes and returns a new instance
// of our custome testServer type
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// initialize a new cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client. Any response cookies will now be stored
	// and sent with subsequent requests when using this client.
	ts.Client().Jar = jar

	// Disable redirect-following for the test server client by setting a custom CheckDirect function.
	// This function will be called whenever a 3xx response is received by the client, and by always reeturning
	// a http.ErrUseLastResponse error it forces the client to immediately return the received response
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}

}

// Implement a get() method on our custom testServer type. This makes a GET request
// to a given url path using the test server client, and returns the response status code
// headers and body
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
