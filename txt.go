package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TxtRecord struct {
	Prefix string
	Type   string
	Value  string
}

func (app *application) printTxtRecords() {
	for _, txt := range app.txts {
		fmt.Println("---")
		fmt.Printf("Prefix:\t%s\nType:\t%s\nValue\t%s\n", txt.Prefix, txt.Type, txt.Value)
		fmt.Println("---")
	}
}

func (app *application) setAcmeTo(valid string) {
	for _, txt := range app.txts {
		if txt.Prefix == "_acme-challenge" {
			txt.Value = valid
			app.infoLog.Println("acme changed")
			return
		}
	}
	txt := TxtRecord{Prefix: "_acme-challenge", Type: "TXT", Value: valid}
	app.txts = append(app.txts, &txt)
	app.infoLog.Println("acme added")
}

func (app *application) postTxtRecords() error {

	data := fmt.Sprintf("sessionID=%s&cID=1&node=ManageDomains&vhost=%s&spf_type=NONE", app.sessionID, app.domain)
	for _, txt := range app.txts {
		tmp := fmt.Sprintf("&prefix=%s&type=%s&value=%s", txt.Prefix, txt.Type, txt.Value)
		app.infoLog.Println(tmp)
		data += tmp
	}
	data += "&action_change_txt_records=Einstellung+%C3%BCbernehmen"

	body := strings.NewReader(data)

	link := fmt.Sprintf("https://www.strato.de/apps/CustomerService?sessionID=%s&cID=1&node=ManageDomains&action_show_txt_records&vhost=%s", app.sessionID, app.domain)
	req, err := http.NewRequest("POST", link, body)
	if err != nil {
		return err
	}
	req.Host = "www.strato.de"
	req.Header.Set("Connection", "close")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Origin", "https://www.strato.de")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/80.0.3987.87 Chrome/80.0.3987.87 Safari/537.36")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	referer := fmt.Sprintf("https://www.strato.de/apps/CustomerService?sessionID=%s&cID=1&node=ManageDomains&action_show_txt_records&vhost=%s", app.sessionID, app.domain)
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	for _, c := range app.cookies {
		req.AddCookie(c)
	}
	resp, err := app.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (app *application) getTxtRecords() error {
	var txts []*TxtRecord

	link := fmt.Sprintf("https://www.strato.de/apps/CustomerService?sessionID=%s&cID=1&node=ManageDomains&action_show_txt_records&vhost=%s", app.sessionID, app.domain)

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}

	req.Host = "www.strato.de"
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	referer := fmt.Sprintf("https://www.strato.de/apps/CustomerService?sessionID=%s&cID=1&node=kds_DomainManagement&action_settings.x=1&domain=%s&id=5&host_id=0", app.sessionID, app.domain)
	req.Header.Set("Referer", referer)
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
	txts, err = parseTxtRecords(resp)
	if err != nil {
		return err
	}
	app.txts = txts
	return nil
}

func parseTxtRecords(resp *http.Response) ([]*TxtRecord, error) {
	var txts []*TxtRecord
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return txts, err
	}

	doc.Find("div.txt-record-tmpl").Each(func(i int, s *goquery.Selection) {
		var txt TxtRecord

		prefix, _ := s.Find("input.form-control").Attr("value")
		txt.Prefix = prefix

		s.Find("option").Each(func(i int, o *goquery.Selection) {
			_, selected := o.Attr("selected")
			if selected {
				txtType, _ := o.Attr("value")
				txt.Type = txtType
			}
		})

		text := s.Find("textarea.form-control").Text()
		txt.Value = text

		txts = append(txts, &txt)
	})
	return txts, nil
}
