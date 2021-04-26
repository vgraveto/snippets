package main

import (
	"context"
	"encoding/gob"
	"flag"
	"github.com/golangcollege/sessions"
	"github.com/vgraveto/snippets/cmd/web/handlers"
	"github.com/vgraveto/snippets/pkg/models"
	"github.com/vgraveto/snippets/pkg/models/dbapi"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	snippetsWebVersion = "snippetsWEB_0_0_4"
	buildCounter       = 8
)

// cmd line parameters
var (
	configPath = flag.String("config", "", "the path for the config file")
	configFile = flag.String("configfile", "snippetsWeb", "the filename of the config file - without required .toml extension")
	certsPath  = flag.String("certs", "certs", "the path for certificates location")
	keysPath   = flag.String("privatekeys", "certs", "the path for private keys location")
	debugOn    = flag.Bool("debug", false, "Enable debug mode")
	secret     = flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Session Manager Secret key")
)

func main() {
	// Use log.New() to create a logger for writing information messages. This takes
	// three parameters: the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "Snippets WEB - INFO\t", log.Ldate|log.Ltime)
	// Create a logger for writing error messages in the same way, but use stderr as
	// the destination and use the log.Lshortfile flag to include the relevant
	// file name and line number.
	errorLog := log.New(os.Stderr, "Snippets WEB - ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Printf("main: Initializing ... %s build %d\n", snippetsWebVersion, buildCounter)
	defer func() {
		infoLog.Printf("main: Terminating ... %s build %d\n", snippetsWebVersion, buildCounter)
	}()

	// Parse de command line parameters
	flag.Parse()
	if len(*certsPath) == 0 {
		*certsPath = "."
	}
	if len(*keysPath) == 0 {
		*keysPath = "."
	}

	// read configuration file from disk
	globalData := readConfig(errorLog, *configPath, *configFile)

	// setup signal catching
	sigs := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C), SIGKILL, SIGQUIT or SIGTERM
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGTERM)

	db, err := dbapi.DialDB(infoLog, globalData.DB)
	if err != nil {
		errorLog.Fatalf("main: %v\n", err)
	}
	defer func() {
		err = dbapi.CloseDB(infoLog, db)
		if err != nil {
			errorLog.Printf("main: Closing api database error: %v\n", err)
			return
		}
	}()

	// Initialize a new template cache...
	templateCache, err := handlers.NewTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	} else {
		infoLog.Println("main: Template parsing is OK")
	}

	// Use the sessions.New() function to initialize a new session manager,
	// passing in the secret key as the parameter. Then we configure it so
	// sessions always expires after sessionLifetime.
	// TODO session is not working well on Safari browser !!!!
	// Register the custom type with the encoding/gob package.
	gob.Register(models.TokenMessage{}) // used on session
	gob.Register(models.TokenUser{})    // used on session as field of models.TokenMessage
	gob.Register(dbapi.UserModel{})     // needed because Users field is of this type on handlers.Application
	gob.Register(dbapi.SnippetModel{})  // needed because Snippets field is of this type on handlers.Application
	// create new session
	session := sessions.New([]byte(*secret))
	session.Lifetime = globalData.sessionLifetime
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	// Initialize a new instance of application containing the dependencies.
	app := &handlers.Application{
		DebugOn:       *debugOn,
		ErrorLog:      errorLog,
		InfoLog:       infoLog,
		Session:       session,
		TemplateCache: templateCache,
		Snippets:      dbapi.NewSnippetModel(db),
		Users:         dbapi.NewUserModel(db),
	}

	httpSrv := &http.Server{
		Addr:         ":" + globalData.HttpPort,
		ErrorLog:     errorLog,
		Handler:      app.Routes(),
		IdleTimeout:  globalData.HttpIdleTimeout,
		ReadTimeout:  globalData.HttpReadTimeout,
		WriteTimeout: globalData.HttpWriteTimeout,
	}

	// Initialize this Web Frontend
	var mainError error

	go func() {
		infoLog.Printf("main: starting http server at %q\n", httpSrv.Addr)

		mainError = httpSrv.ListenAndServe()
		if mainError != nil {
			errorLog.Printf("main: httpSrv.ListenAndServe failed: %s\n", mainError)
			sigs <- os.Interrupt
		} else {
			infoLog.Printf("main: httpSrv.ListenAndServe stopped with no errors\n")
		}
	}()

	// Block main until a signal is received
	s := <-sigs

	infoLog.Printf("main: received signal: %s", s)
	AppCleanup(infoLog, httpSrv, globalData.deadlineWaitForClose)

}

func AppCleanup(infoLog *log.Logger, httpSrv *http.Server, wait time.Duration) {
	infoLog.Println("main: AppCleanup init")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	httpSrv.Shutdown(ctx)

	infoLog.Println("main: AppCleanup end")
}
