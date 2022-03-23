package uviews

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/usfsci/ustore"
	"golang.org/x/text/message"
)

const (
	apiVersion              = "0.00"
	contentTypeKey          = "Content-Type"
	contentDispositionKey   = "Content-Disposition"
	acceptLanguageHeaderKey = "Accept-Language"

	debugMode = true

	//contentValueDispositionAttachement = "attachment"
	//contentValueOctetStream            = "application/octet-stream"
)

type App struct {
	Router  *mux.Router
	port    string
	dataDir string
	// Either http or https
	protocol string
	// App Name used to derive Cookies name's
	name string
	// CSRF Secret Key
	csrfKey []byte
	// Requests with no auth are redirected here
	notAuthPath string
}

// NewApp - Creates and configures Router
// authKey: CSRF Secret Key
// port: Server port
// dataDir: Directory for Data
// rootPath: URL path for calls to root
func NewApp(appName string, csrfKey []byte, port string, dataDir string, rootPath string, notAuthPath string) *App {
	appName = "_" + strings.TrimSpace(strings.ToLower(appName))

	// Sets up the session ID
	SetupSessions(appName + "sessionid")

	r := mux.NewRouter().StrictSlash(true)

	// File server for static content
	fserver := http.FileServer(http.Dir(dataDir + "/static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fserver))

	app := &App{
		Router:      r,
		dataDir:     dataDir,
		port:        port,
		notAuthPath: notAuthPath,
		name:        appName,
		csrfKey:     csrfKey,
	}

	// CSRF middleware
	/*csrfMiddleware := csrf.Protect(
		csrfKey,
		csrf.Secure(true),
		csrf.CookieName(appName+"csfr"),
		csrf.HttpOnly(true),
		csrf.Path("/"),
		csrf.MaxAge(6*30*86400), // 6 month to avoid overnigth issues
	)
	log.Printf("MaxAge set\n")*/

	// Enable middlewares
	r.Use(loggingMiddleware)
	r.Use(app.redirectMiddleware)
	//r.Use(csrfMiddleware)

	// Return an instance of the App
	return app
}

func (app *App) EnableCSRF() {
	app.Router.Use(
		csrf.Protect(
			app.csrfKey,
			csrf.Secure(true),
			csrf.CookieName(app.name+"csfr"),
			csrf.HttpOnly(true),
			csrf.Path("/"),
			csrf.MaxAge(6*30*86400), // 6 month to avoid overnigth issues
		),
	)
}

func (app *App) RunApp() {
	// Start server
	unsecureServerStart(app.Router, app.port)
}

// unsecureServerStart - Start server on HTTP mode
func unsecureServerStart(r *mux.Router, port string) {
	log.Printf("start http server on port %s.", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("could not start HTTP server on port %s (%s)\n", port, err)
	}
}

/*// secureServerStart - Start server on HTTPS mode
func secureServerStart(r *mux.Router) {
	if err := http.ListenAndServeTLS(
		":"+config.LyceumProperties.ServerConf.ServerPort,
		config.LyceumProperties.CertConf.Path,
		config.LyceumProperties.CertConf.KeyPath,
		r,
	); err != nil {
		log.Fatalf("could not start HTTPS server on port %s (%s)\n", config.LyceumProperties.ServerConf.ServerPort, err)
	}
}*/

// loggingMiddleware - Logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("%s %s: %s %s\n%+v\n", r.RemoteAddr, r.Proto, r.Method, r.RequestURI, r.Header)
		log.Printf("\n%+v\n\n", r.Host)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// redirectMiddleware - Ensures that requests that use a prefix are redirected to the
// Canonical server.
func (app *App) redirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.Split(r.Host, ".")
		if len(host) == 3 {
			rd := fmt.Sprintf(
				"%s://%s.%s%s",
				app.protocol,
				host[1],
				host[2],
				r.URL.Path)
			http.Redirect(w, r, rd, http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ViewGetHandler - Wrapper for view GET
// If the user is not entitled to read from this view
// redirects to notAuthURL
func (app *App) ViewGetHandler(w http.ResponseWriter, r *http.Request, v View) {
	// Authorize
	ok, err := v.CanRead(r.Context(), mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirect(w, r, app.notAuthPath, http.StatusSeeOther)
		return
	}

	v.Get(w, r)
}

// ViewPostHandler - Decodes POST form and calls view post
// Verifies that the user is entitled to write to this view
func (app *App) ViewPostHandler(w http.ResponseWriter, r *http.Request, v View) {
	defer r.Body.Close()

	// Check authorization
	ok, err := v.CanWrite(r.Context(), mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirect(w, r, app.notAuthPath, http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//log.Printf("POST:\n%+v\n", r.Form)

	// Delete the CSRF token to avoid Decoder error
	delete(r.PostForm, "gorilla.csrf.Token")

	v.Post(w, r)
}

// ViewPostHandler - Decodes POST form and calls view post
// Verifies that the user is entitled to write to this view
func (app *App) RestPostHandler(w http.ResponseWriter, r *http.Request, entity ustore.Entity, u *ustore.User) {
	defer r.Body.Close()

	// TODO: Check authorization
	//ok, err := entity.CanWrite(u)
	/*ok, err := entity.IsAuthorized(r.Context(), u, )
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirect(w, r, app.notAuthPath, http.StatusSeeOther)
		return
	}*/

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//log.Printf("POST:\n%+v\n", r.Form)

	// Delete the CSRF token to avoid Decoder error
	delete(r.PostForm, "gorilla.csrf.Token")

	//v.Post(w, r)
}

func (app *App) Authenticate(newView func() View, viewHandler func(http.ResponseWriter, *http.Request, View)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view := newView()
		if err := view.Load(w, r); err != nil {
			return
		}

		if view.GetUser() == nil || !view.GetUser().EmailConfirmed {
			redirect(w, r, app.notAuthPath, http.StatusSeeOther)
			return
		}

		viewHandler(w, r, newView())
	}
}

func (app *App) BypassAuthentication(newView func() View, viewHandler func(http.ResponseWriter, *http.Request, View)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view := newView()
		if err := view.Load(w, r); err != nil {
			return
		}

		viewHandler(w, r, view)
	}
}

// getLanguage - Parses the Accept Language header to select proper localization
func getLanguage(r *http.Request) string {
	langList := r.Header.Get(acceptLanguageHeaderKey)

	// Defaul to US English
	lang := "en-US"
	if len(langList) > 0 {
		lang = strings.Split(langList, ";")[0]
	}

	return lang
}

func templateFuncs(lang string) template.FuncMap {
	return template.FuncMap{
		"pPrintf": message.NewPrinter(message.MatchLanguage(lang)).Sprintf,
		"inc":     func(i int) int { return i + 1 },
		"fieldState": func(key string, missing string) string {
			if key == missing {
				return invalidInputFlag
			}
			return ""
		},
	}
}
