package dbmysql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"time"
)

const (
	// name used to register a new TLS configuration
	tlsConfigName = "dbaseTLSconfig"
)

type DBdata struct {
	Protocol          string
	Server            string
	Dbase             string
	Username          string
	Password          string
	ServerCA          string
	ClientCert        string
	ClientKey         string
	DbConnMaxLifetime time.Duration // number of seconds
}

func RefreshDBConnection(infoLog *log.Logger, caller string, db *sql.DB) (err error) {
	if err = db.Ping(); err != nil {
		return fmt.Errorf("RefreshDBConnection: RefreshDBConnection by %q: %v", caller, err)
	}
	infoLog.Printf("RefreshDBConnection: by %q\n", caller)
	return nil
}

func DialDB(infoLog *log.Logger, dialData DBdata, certsPath, keysPath string) (db *sql.DB, err error) {
	infoLog.Println("DialDB: loading certificates")

	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(certsPath + dialData.ServerCA)
	if err != nil {
		return nil, fmt.Errorf("DialDB: %v\n", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return nil, fmt.Errorf("dbase: Failed to append server PEM.")
	}
	clientCert := make([]tls.Certificate, 0, 1)
	if len(dialData.ClientCert) == 0 || len(dialData.ClientKey) == 0 {
		infoLog.Printf("dbase: No client certificate specified - only server side certificate will be used\n")
		clientCert = nil
	} else {
		certs, err := tls.LoadX509KeyPair(certsPath+dialData.ClientCert, keysPath+dialData.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("dbase: %v\n", err)
		}
		clientCert = append(clientCert, certs)
	}
	err = mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
		RootCAs:            rootCertPool,
		Certificates:       clientCert,
		InsecureSkipVerify: true, // TODO lack of security accept any certificate and host name provided by the server
	})
	if err != nil {
		return nil, fmt.Errorf("DialDB: RegisterTLSConfig: %v\n", err)
	}

	infoLog.Println("DialDB: dialing mysql database")

	cfg := mysql.NewConfig()
	cfg.Net = dialData.Protocol
	cfg.Addr = dialData.Server
	cfg.User = dialData.Username
	cfg.Passwd = dialData.Password
	cfg.DBName = dialData.Dbase
	cfg.TLSConfig = tlsConfigName
	cfg.ParseTime = true
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("DialDB: %v\n", err)
	}
	db.SetConnMaxLifetime(dialData.DbConnMaxLifetime) // imposed for correct work of reconnection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("DialDB: error on Ping: %v\n", err)
	}

	infoLog.Println("DialDB: connection to database is OK")
	return db, err
}

func CloseDB(infoLog *log.Logger, db *sql.DB) error {
	err := db.Close()
	if err != nil {
		return fmt.Errorf("CloseDB: CloseDB: %v", err)
	}
	infoLog.Println("CloseDB: closed dbmysql connection")
	return err
}
