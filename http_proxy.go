package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// consume list of templates and release their full path counterparts
func fileNames(filenames ...string) []string {
	tdir := "./" // template directory
	flist := []string{}
	for _, fname := range filenames {
		flist = append(flist, tdir+fname)
	}
	return flist
}

// helper function to close the file
func closeFile(fd *os.File) {
	// close fd on exit and check for its returned error
	err := fd.Close()
	checkError(err)
}

// helper function to check the error
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// read txt file and return list of lines separated by newline
func readTxtFile(fname string) []string {
	var out []string
	// open file and read its content into byte's buffer
	fi, err := os.Open(fname)
	if err != nil {
		return out
	}
	defer closeFile(fi)
	checkError(err)
	// make a reader buffer
	rb := bufio.NewReader(fi)
	for {
		// read line by line
		data, _, err := rb.ReadLine()
		if err != nil {
			break
		}
		line := string(data)
		// skip line which starts with #
		if len(line) > 1 && strings.HasPrefix(line, "#") == true {
			continue
		}
		out = append(out, line)
	}
	return out
}

// read CSV file and return headers and values
func readCSVFile(fname string) [][]string {
	// open file and read its content into byte's buffer
	fi, err := os.Open(fname)
	defer closeFile(fi)
	checkError(err)
	// make a reader buffer
	rb := bufio.NewReader(fi)
	// make CSV reader
	r := csv.NewReader(rb)
	out, err := r.ReadAll()
	checkError(err)
	return out
}

type Rules struct {
	Url     string
	MinHour int
	MaxHour int
}

func parseRules(records [][]string) []Rules {
	var rules []Rules
	var r Rules
	for _, row := range records {
		r.Url = row[0]
		r.MinHour, _ = strconv.Atoi(row[1])
		r.MaxHour, _ = strconv.Atoi(row[2])
		rules = append(rules, r)
	}
	return rules
}

func myproxy() {

	var port, wlistFile, blistFile, ruleFile string
	var verbose int
	flag.StringVar(&port, "port", ":9999", "Proxy port number")
	flag.StringVar(&wlistFile, "whitelist", "whitelist.txt", "White list file")
	flag.StringVar(&blistFile, "blacklist", "blacklist.txt", "Black list file")
	flag.StringVar(&ruleFile, "rules", "rules.txt", "Rule list file")
	flag.IntVar(&verbose, "verbose", 0, "logging level")
	flag.Parse()

	// init proxy server
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false // true
	msg := fmt.Sprintf("port=%s, verbose=%d, wlist=%s, blist=%s, rule=%s", port, verbose, wlistFile, blistFile, ruleFile)
	log.Println(msg)

    // read out client settings
	whitelist := readTxtFile(wlistFile)
	log.Println("White list:", whitelist)
	blacklist := readTxtFile(blistFile)
	log.Println("Black list:", blacklist)
	rulelist := parseRules(readCSVFile(ruleFile))
	log.Println("Rule list:", rulelist)

	// admin handler
	proxy.OnRequest(goproxy.DstHostIs("")).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			tpage := "admin.tmpl.html"
			data := map[string]interface{}{
				"whitelist": whitelist,
				"blacklist": blacklist,
				"rulelist":  rulelist,
			}
			buf := new(bytes.Buffer)
			filenames := fileNames(tpage)
			t := template.Must(template.ParseFiles(filenames...))
			err := t.Execute(buf, data)
			checkError(err)
			return r, goproxy.NewResponse(r, goproxy.ContentTypeHtml,
				http.StatusOK, buf.String())
		})

	// restrict certain sites on time based rules
	for _, rule := range rulelist {
		proxy.OnRequest(goproxy.DstHostIs(rule.Url)).DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				h, _, _ := time.Now().Clock()
				if h < rule.MinHour && h > rule.MaxHour {
					return r, goproxy.NewResponse(r,
						goproxy.ContentTypeText, http.StatusForbidden,
						"Your exceed your time window on this site")
				}
				return r, nil
			})
	}

	// filter white/black lists
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			pat1 := strings.Join(whitelist, "$|")
			expect1 := false // match=false means site not in whitelist
			match1, err := regexp.MatchString(pat1, r.URL.Host)
			if err != nil {
				log.Println("ERROR: fail in match", pat1, r.URL.Host)
			}
			pat2 := strings.Join(blacklist, "$|")
			expect2 := true // match=true means site is in blacklist
			match2, err := regexp.MatchString(pat2, r.URL.Host)
			if err != nil {
				log.Println("ERROR: fail in match", pat2, r.URL.Host)
			}
			if match1 == expect1 && match2 == expect2 {
				path := html.EscapeString(r.URL.Path)
				if verbose > 0 {
					log.Println(r.URL.Host, path)
				}
				return r, goproxy.NewResponse(r,
					goproxy.ContentTypeText, http.StatusForbidden,
					"This site is not accessible to you")
			}
			return r, nil
		})

	// start and log http proxy server
	log.Fatal(http.ListenAndServe(port, proxy))

}

func main() {
	myproxy()
}
