package main

import (
	"context"
	"flag"
	"github.com/vgraveto/snippets/cmd/api/handlers"
	"github.com/vgraveto/snippets/pkg/models"
	"github.com/vgraveto/snippets/pkg/models/dbmysql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	snippetsApiVersion = "snippetsAPI_0_0_6"
	buildCounter       = 8
)

// cmd line parameters
var (
	configPath = flag.String("config", "", "the path for the config file")
	configFile = flag.String("configfile", "snippetsApi", "the filename of the config file - without required .toml extension")
	certsPath  = flag.String("certs", "certs", "the path for certificates location")
	keysPath   = flag.String("privatekeys", "certs", "the path for private keys location")
	debugOn    = flag.Bool("debug", false, "Enable debug mode")
)

func main() {
	// Use log.New() to create a logger for writing information messages. This takes
	// three parameters: the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "Snippets API - INFO\t", log.Ldate|log.Ltime)
	// Create a logger for writing error messages in the same way, but use stderr as
	// the destination and use the log.Lshortfile flag to include the relevant
	// file name and line number.
	errorLog := log.New(os.Stderr, "Snippets API - ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Printf("main: Initializing ... %s build %d\n", snippetsApiVersion, buildCounter)
	defer func() {
		infoLog.Printf("main: Terminating ... %s build %d\n", snippetsApiVersion, buildCounter)
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

	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate dialDB() function below.
	db, err := dbmysql.DialDB(infoLog, globalData.DB, *certsPath+"/", *keysPath+"/")
	if err != nil {
		errorLog.Fatalf("main: %v\n", err)
	}
	// We also defer a call to CloseDB(), so that the connection pool is closed
	// before the main() function exits.
	defer func() {
		err = dbmysql.CloseDB(infoLog, db)
		if err != nil {
			errorLog.Printf("main: Closing database error: %v\n", err)
			return
		}
	}()

	// Initialize a new instance of application containing the dependencies.
	app := &handlers.Application{
		DebugOn:  *debugOn,
		ErrorLog: errorLog,
		InfoLog:  infoLog,
		Snippets: dbmysql.NewSnippetModel(db),
		Users:    dbmysql.NewUserModel(db),
		Tokens:   models.NewTokenModel(&globalData.TD),
		Val:      models.NewValidation(),
	}

	httpSrv := &http.Server{
		Addr:         ":" + globalData.HttpPort,
		ErrorLog:     errorLog,
		Handler:      app.Routes(),
		IdleTimeout:  globalData.HttpIdleTimeout,
		ReadTimeout:  globalData.HttpReadTimeout,
		WriteTimeout: globalData.HttpWriteTimeout,
	}

	// Initialize this REST API
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
