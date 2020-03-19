package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (app *application) prepareApp() error {
	err := app.getStratoProSession()
	if err != nil {
		return err
	}
	err = app.getCookieCookie()
	if err != nil {
		return err
	}

	err = app.getKsbCookie()
	if err != nil {
		return err
	}

	err = app.getSessionId()
	if err != nil {
		return err
	}

	return nil
}

func (app *application) getStratoProSession() error {

	link := "https://www.strato.de/buy/ger/basket/count?callback=basketcount&_=" + fmt.Sprintf("%d", makeTimestamp())

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}
	req.Host = "www.strato.de"
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0")
	req.Header.Set("Accept", "text/javascript, application/javascript, application/ecmascript, application/x-ecmascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "close")
	req.Header.Set("Referer", "https://www.strato.de/")

	resp, err := app.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	app.cookies = append(app.cookies, resp.Cookies()...)
	return nil
}

func (app *application) getCookieCookie() error {

	req, err := http.NewRequest("POST", "https://orca.strato.de/orca/trkn/ger/1111", nil)
	if err != nil {
		return err
	}
	req.Host = "orca.strato.de"
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Origin", "https://www.strato.de")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "close")
	req.Header.Set("Referer", "https://www.strato.de/")
	req.Header.Set("Content-Length", "0")
	resp, err := app.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	app.cookies = append(app.cookies, resp.Cookies()...)

	return nil
}

func (app *application) getKsbCookie() error {
	req, err := http.NewRequest("GET", "https://www.strato.de/apps/CustomerService", nil)
	if err != nil {
		return err
	}

	req.Host = "www.strato.de"
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "close")
	req.Header.Set("Referer", "https://www.strato.de/")

	for _, c := range app.cookies {
		req.AddCookie(c)
	}
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := app.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	app.cookies = append(app.cookies, resp.Cookies()...)

	return err
}

func (app *application) getSessionId() error {

	data := fmt.Sprintf("identifier=%s&passwd=%s&action_customer_login.x=Login", app.config.Login, app.config.Password)
	body := strings.NewReader(data)
	req, err := http.NewRequest("POST", "https://www.strato.de/apps/CustomerService", body)
	if err != nil {
		return err
	}
	req.Host = "www.strato.de"
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Referer", "https://www.strato.de/apps/CustomerService")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "83")
	req.Header.Set("Origin", "https://www.strato.de")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "close")
	for _, c := range app.cookies {
		req.AddCookie(c)
	}
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := app.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	location := resp.Header["Location"]
	if len(location) == 0 {
		return errors.New("most probably password is wrong, but could be something else")
	}
	u, err := url.Parse(location[0])
	if err != nil {
		return err
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}
	if len(m) == 0 {
		return errors.New("something went wrong, strato.de did not return redirect url, hmmm")
	}
	app.sessionID = m["sessionID"][0]
	return nil
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
