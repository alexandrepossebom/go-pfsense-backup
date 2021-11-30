package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"go-pfsense-backup/config"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Repository struct
type Repository struct {
	jar   http.CookieJar
	conf  config.FirewallItem
	data  url.Values
	debug bool
}

func (r *Repository) post() string {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Jar:       r.jar,
	}
	request, err := http.NewRequest("POST", r.conf.URL+"/diag_backup.php", strings.NewReader(r.data.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Fatal(err.Error())
	}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(body)
}

func (r *Repository) download() string {
	r.jar, _ = cookiejar.New(nil)
	r.data = url.Values{}
	// Init
	re := regexp.MustCompile(`name=\'__csrf_magic\'\s*value="([^"]+)"`)
	parts := re.FindStringSubmatch(r.post())
	if len(parts) != 2 {
		log.Fatal("invalid csrf")
	}
	csrf := parts[1]

	// Login
	r.data = url.Values{}
	r.data.Add("__csrf_magic", csrf)
	r.data.Add("usernamefld", r.conf.Username)
	r.data.Add("passwordfld", r.conf.Password)
	r.data.Add("login", "Login")

	login := r.post()
	if strings.Contains(login, "Username or Password incorrect") {
		log.Fatal("Invalid password")
	}
	re = regexp.MustCompile(`name=\'__csrf_magic\'\s*value="([^"]+)"`)
	parts = re.FindStringSubmatch(login)

	if len(parts) != 2 {
		log.Fatal("invalid csrf")
	}
	csrf = parts[1]

	// Download
	r.data = url.Values{}
	r.data.Add("__csrf_magic", csrf)
	r.data.Add("download", "download")
	r.data.Add("Submit", "download") // for compatibility with old pfsense versions
	r.data.Add("donotbackuprrd", "yes")
	xml := r.post()
	if !strings.Contains(xml, `<?xml version="1.0"?>`) {
		if r.debug {
			fmt.Println(xml)
		}
		log.Fatalln("XML file is invalid.")
	}
	return xml
}

func (r *Repository) write(txt string) error {
	if _, err := os.Stat(r.conf.Directory); os.IsNotExist(err) {
		log.Fatalln("Directory not exists: " + r.conf.Directory)
	}
	filename := fmt.Sprintf("%s-%s.xml", r.conf.Name, time.Now().Format("2006-01-02"))
	filename = filepath.Join(r.conf.Directory, filename)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.WriteString(txt)
	if err != nil {
		log.Fatal(err)
	}
	return file.Close()
}

func main() {
	debug := flag.Bool("debug", false, "debug site output if errors")
	flag.Parse()
	for _, firewall := range config.Get().Firewalls {
		start := time.Now()
		fmt.Printf("Starting backup of %s ...", firewall.Name)
		r := &Repository{conf: firewall, debug: *debug}
		if err := r.write(r.download()); err != nil {
			log.Fatalln("Error writing file.")
		}
		duration := time.Since(start)
		fmt.Printf("done %s\n", duration)
	}
	fmt.Println("Done")
}
