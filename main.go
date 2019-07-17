package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type content struct {
	Str string `json:"content"`
}

type file struct {
	Desc  string             `json:"description,omitempty"`
	Pub   bool               `json:"public"`
	Files map[string]content `json:"files"`
}

type url struct {
	URL string `json:"html_url"`
}

const envar = "GISTAUTH"
var auth = os.Getenv(envar)

func main() {
	var f file
	flag.StringVar(&f.Desc, "m", "", "gist description")
	flag.BoolVar(&f.Pub, "p", false, "create public gist")
	flag.Usage = usage
	flag.Parse()
	log.SetPrefix("gist: ")
	log.SetFlags(0)

	upass := strings.Split(":", auth)
	if len(upass) < 2 {
		log.Printf("env: %q expected user:secret, got: %q", envar, auth)
		flag.Usage()
		os.Exit(1)
	}

	f.Files = make(map[string]content)
	if len(flag.Args()) == 0 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		f.Files["<stdin>"] = content{string(b)}
		goto Do
	}
	for _, name := range flag.Args() {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			log.Println("gist:", err)
			continue
		}
		f.Files[name] = content{string(b)}
	}
Do:
	js, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
	}
	const api = `https://api.github.com/gists`
	req, err := http.NewRequest("POST", api, bytes.NewReader(js))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(upass[0], upass[1])

	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var inc url
	err = json.NewDecoder(resp.Body).Decode(&inc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(inc.URL)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gist [ options ] [ file... ]\n")
	flag.PrintDefaults()
}
