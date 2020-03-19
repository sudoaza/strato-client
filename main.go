package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
)

type Config struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type application struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	config    Config
	client    *http.Client
	domain    string
	cookies   []*http.Cookie
	sessionID string
	txts      []*TxtRecord
}

func main() {
	domain := flag.String("d", "", "domain to set txt record to")
	valid := flag.String("valid", "", "valid string")
	path := flag.String("c", "config.yml", "path to the config file")
	flag.Parse()

	c, err := parseConfig(*path)
	if err != nil {
		log.Println(err)
		return
	}

	if *domain == "" {
		flag.Usage()
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	client := setHttpClient()
	app := application{
		client:   client,
		domain:   *domain,
		config:   c,
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	err = app.prepareApp()
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	err = app.getTxtRecords()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	if *valid != "" {
		app.setAcmeTo(*valid)
		err = app.postTxtRecords()
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	} else {
		app.printTxtRecords()
	}
}

func parseConfig(p string) (Config, error) {
	c := Config{}
	rawConfig, err := ioutil.ReadFile(p)
	if err != nil {
		flag.Usage()
		return c, err
	}
	err = yaml.Unmarshal(rawConfig, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func setHttpClient() *http.Client {
	// tr := &http.Transport{
	// 	Proxy:           http.ProxyFromEnvironment,
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// Transport: tr,
	}
}
