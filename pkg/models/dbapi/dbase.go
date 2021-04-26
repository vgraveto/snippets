package dbapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type DBapi struct {
	URL string // the URL where the middleware API is deployed
}

type API struct {
	Url string
}

func DialDB(infoLog *log.Logger, dialData DBapi) (db *API, err error) {
	if dialData.URL == "" {
		return nil, fmt.Errorf("dbapi: DialDB: inalid URL - %q", dialData.URL)
	}
	resp, err := http.Get(dialData.URL)
	if err != nil {
		return nil, fmt.Errorf("DialDB: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("DialDB: ReadAll: %v", err)
		}
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("DialDB: Status: %s - %s", resp.Status, bodyString)
	}
	infoLog.Printf("DialDB: API database checked")
	// valid URL and API running
	return &API{Url: dialData.URL}, nil
}

func CloseDB(infoLog *log.Logger, db *API) error {
	db.Url = ""
	return nil
}
