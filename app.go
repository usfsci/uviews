package uviews

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/usfsci/ustore"
	"golang.org/x/text/message"
)

const (
	contentTypeKey          = "Content-Type"
	contentDispositionKey   = "Content-Disposition"
	acceptLanguageHeaderKey = "Accept-Language"

	contentValueDispositionAttachement = "attachment"
	contentValueOctetStream            = "application/octet-stream"
)

type App struct {
	Router  *mux.Router
	port    string
	dataDir string
	// Either http or https
	protocol string
}

// NewApp - Creates and configures Router
// authKey: CSRF Secret Key
// port: Server port
// dataDir: Directory for Data
// rootPath: URL path for calls to root
func NewApp(appName string, authKey []byte, port string, dataDir string, rootPath string) *App {
	appName = "_" + strings.TrimSpace(strings.ToLower(appName))

	// Sets up the session ID
	SetupSessions(appName + "sessionid")

	r := mux.NewRouter().StrictSlash(true)

	// File server for static content
	fserver := http.FileServer(http.Dir(dataDir + "/static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fserver))

	// Login handlers
	/*r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, rootPath, http.StatusSeeOther)
	})*/

	// TODO: Install route handlers

	// CSRF middleware
	csrfMiddleware := csrf.Protect(
		authKey,
		csrf.Secure(true),
		csrf.CookieName(appName+"csfr"),
		csrf.HttpOnly(true),
		csrf.Path("/"),
		csrf.MaxAge(6*30*86400), // 6 month to avoid overnigth issues
	)
	log.Printf("MaxAge set\n")

	app := &App{
		Router:  r,
		dataDir: dataDir,
		port:    port,
	}

	// Enable middlewares
	r.Use(loggingMiddleware)
	r.Use(app.redirectMiddleware)
	r.Use(csrfMiddleware)

	// Return an instance of the App
	return app
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

func (app *App) sessionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

// ViewGetHandler - Wrapper for view GET
// Sets the view session and user
// Verifies that the user is entitled to read from this view
func ViewGetHandler(w http.ResponseWriter, r *http.Request, s *ustore.Session, v View) {
	// Update view with Session & User
	v.SetSession(s)

	// Authorize
	ok, err := v.CanRead(r.Context(), mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirect(w, r, "/ipc/login", http.StatusSeeOther)
		return
	}

	v.Get(w, r)
}

// ViewPostHandler - Decodes POST form and calls view post
// Sets the view session and user
// Verifies that the user is entitled to write to this view
func ViewPostHandler(w http.ResponseWriter, r *http.Request, s *ustore.Session, v View) {
	defer r.Body.Close()

	// Update view with Session information
	v.SetSession(s)

	// Check authorization
	ok, err := v.CanWrite(r.Context(), mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		redirect(w, r, "/ipc/login", http.StatusSeeOther)
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

func RenderForm(w http.ResponseWriter, r *http.Request, templateFiles []string, session *ustore.Session, f Form) {
	lang := getLanguage(r)

	t, err := template.New(filepath.Base(templateFiles[0])).Funcs(templateFuncs(lang)).ParseFiles(templateFiles...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always add CSRF to any Forms
	f.SetCsrf(csrf.TemplateField(r))

	f.SetLoggedIn(session != nil && session.User != nil)

	// Serve the template
	if err := t.Execute(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Authenticate(newView func() View, viewHandler func(http.ResponseWriter, *http.Request, *ustore.Session, View)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := LoadSession(w, r)
		if err != nil || session == nil {
			redirect(w, r, "/ipc/login", http.StatusSeeOther)
			return
		}

		if session.User == nil || !session.User.TokenConfirmed {
			redirect(w, r, "/ipc/login", http.StatusSeeOther)
			return
		}

		viewHandler(w, r, session, newView())
	}
}

func BypassAuth(newView func() View, viewHandler func(http.ResponseWriter, *http.Request, *ustore.Session, View)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to get the session
		session, err := LoadSession(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If there was no session set one with UserID nil
		if session == nil {
			session, err = InitSession(w, r, nil)
			if err != nil {
				return
			}
		}

		viewHandler(w, r, session, newView())
	}
}

func loadUser(ctx context.Context, userID ustore.SIDType) (*ustore.User, error) {
	u := &ustore.User{Base: ustore.Base{ID: userID}}
	if err := u.Get(ctx, nil); err != nil {
		return nil, err
	}

	return u, nil
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
