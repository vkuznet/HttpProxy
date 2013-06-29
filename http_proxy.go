package main

import (
	"bufio"
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
	"time"
    "bytes"
    "io/ioutil"
    "text/template"
    "path/filepath"
)

// consume list of templates and release their full path counterparts
func fileNames(tdir string, filenames ...string) []string {
	flist := []string{}
	for _, fname := range filenames {
		flist = append(flist, filepath.Join(tdir, fname))
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
		line := strip(string(data))
		// skip line which starts with #
        if len(line) == 0 {
            continue
        } else {
            if strings.HasPrefix(line, "#") == true {
                continue
            }
            if strings.HasPrefix(line, " ") == true {
                continue
            }
        }
		out = append(out, line)
	}
	return out
}

// function which strip off whitespaces from both side of given string
func strip(s string) string {
    slist := strings.Split(s, "")
    i := -1
    j := -1
    for idx, v := range slist {
        if i == -1 && v != " " {
            i = idx
        }
        back_idx := len(s)-1-idx
        if j == -1 && slist[back_idx] != " " {
            j = back_idx
        }
    }
    if i > 0 && j > 0 {
        return s[i:j+1]
    }
    return s
}

// read CSV file and return headers and values
func readCSVFile(fname string) [][]string {
    var out [][]string
	// open file and read its content into byte's buffer
	fi, err := os.Open(fname)
	if err != nil {
		return out
	}
	defer closeFile(fi)
	checkError(err)
	// make a reader buffer
	rb := bufio.NewReader(fi)
	// make CSV reader
	r := csv.NewReader(rb)
	out, err = r.ReadAll()
	checkError(err)
	return out
}

// save given string to the file
func saveList(fname string, c string) error {
    byteArray := []byte(c)
    err := ioutil.WriteFile(fname, byteArray, 0644)
    return err
}

// rules structure and its methods
type Rules struct {
	Url     string
	MinHour int
	MaxHour int
}
func (r *Rules) ToCSV() string {
    return fmt.Sprintf("%s,%d,%d", r.Url, r.MinHour, r.MaxHour)
}

// helper function to parse given set of records and return list of rules
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

// parse template with given data
func parseTmpl(tdir, tmpl string, data interface{}) string {
    buf := new(bytes.Buffer)
    filenames := fileNames(tdir, tmpl)
    t := template.Must(template.ParseFiles(filenames...))
    err := t.Execute(buf, data)
    checkError(err)
    return buf.String()
}

func myproxy() {

    pname := "HttpProxy4U"
    pver := "1.0.0"

    // get current working directory
    cwd, err := os.Getwd()
    checkError(err)

    // parse input parameters
	var port, wlistFile, blistFile, rulesFile, aname, apwd, tdir, rdir string
	var verbose int
    var interval int64
	flag.StringVar(&port, "port", ":9998", "Proxy port number")
	flag.StringVar(&tdir, "tmpl-dir",
            filepath.Join(cwd, "static/tmpl"), "Template directory")
	flag.StringVar(&rdir, "rule-dir",
            filepath.Join(cwd, "static/rules"), "Rules directory")
	flag.StringVar(&wlistFile, "whitelist",
            filepath.Join(rdir, "whitelist.txt"), "White list file")
	flag.StringVar(&blistFile, "blacklist",
            filepath.Join(rdir, "blacklist.txt"), "Black list file")
	flag.StringVar(&rulesFile, "rules",
            filepath.Join(rdir, "rules.txt"), "Rules list file")
	flag.IntVar(&verbose, "verbose", 0, "logging level")
	flag.Int64Var(&interval, "interval", 300, "reload interval")
	flag.StringVar(&aname, "login", "admin", "Admin login name")
	flag.StringVar(&apwd, "password", "test", "Admin password")
	flag.Parse()

	// init proxy server
	proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = false // true
    if verbose > 1 {
        proxy.Verbose = true
    }
	msg := fmt.Sprintf("port=%s, verbose=%d, wlist=%s, blist=%s, rule=%s",
            port, verbose, wlistFile, blistFile, rulesFile)
	log.Println(msg)

    // read out client settings
	whitelist := readTxtFile(wlistFile)
	log.Println("White list:", whitelist)
	blacklist := readTxtFile(blistFile)
	log.Println("Black list:", blacklist)
	ruleslist := parseRules(readCSVFile(rulesFile))
	log.Println("Rules list:", ruleslist)
    lastRead := time.Now().UTC().Unix()

    // local variables
    var rules []string
    for _, r := range ruleslist {
        rules = append(rules, r.ToCSV())
    }
    tcss := "main.tmpl.css"
    tmplData := map[string]interface{}{}
    css := parseTmpl(tdir, tcss, tmplData)
    tfooter := "footer.tmpl.html"
    tmplData["package"] = pname
    tmplData["version"] = pver
    tmplData["css"] = css
    footer := parseTmpl(tdir, tfooter, tmplData)
    tmplData["whitelist"] = strings.Join(whitelist, "\n")
    tmplData["blacklist"] = strings.Join(blacklist, "\n")
    tmplData["ruleslist"] = strings.Join(rules, "\n")
    tmplData["footer"] = footer

	// admin handler
//    proxy.OnRequest(goproxy.IsLocalHost).DoFunc(
    proxy.OnRequest(goproxy.DstHostIs("")).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
            path := html.EscapeString(r.URL.Path)
            log.Println("admin interface", path)
            if path == "/admin" {
                tauth := "auth.tmpl.html"
                u := r.FormValue("login")
                p := r.FormValue("password")
                if u == aname && p == apwd {
                    log.Println("access admin interface")
                } else {
                    return r, goproxy.NewResponse(r, goproxy.ContentTypeHtml,
                        http.StatusOK, parseTmpl(tdir, tauth, tmplData))
                }
                tpage := "admin.tmpl.html"
                return r, goproxy.NewResponse(r, goproxy.ContentTypeHtml,
                    http.StatusOK, parseTmpl(tdir, tpage, tmplData))
            } else if path == "/save" {
                wlist := strip(r.FormValue("whitelist"))
                blist := strip(r.FormValue("blacklist"))
                rlist := strip(r.FormValue("ruleslist"))
                saveList(wlistFile, wlist)
                saveList(blistFile, blist)
                saveList(rulesFile, rlist)
                tmplData["whitelist"] = wlist
                tmplData["blacklist"] = blist
                tmplData["ruleslist"] = rlist
                tpage := "save.tmpl.html"
                tmplData["body"] = fmt.Sprintf("New rules has been saved on %s",
                        time.Now())
                return r, goproxy.NewResponse(r, goproxy.ContentTypeHtml,
                    http.StatusOK, parseTmpl(tdir, tpage, tmplData))
            } else {
                tpage := "index.tmpl.html"
                return r, goproxy.NewResponse(r, goproxy.ContentTypeHtml,
                    http.StatusOK, parseTmpl(tdir, tpage, tmplData))
            }
		})

	// restrict certain sites on time based rules
	for _, rule := range ruleslist {
		proxy.OnRequest(goproxy.DstHostIs(rule.Url)).DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
                // reload maps if necessary
                unix := time.Now().UTC().Unix()
                if  unix-lastRead > interval {
                    ruleslist := parseRules(readCSVFile(rulesFile))
                    if verbose > 0 {
                        log.Println("Rules list:", ruleslist)
                    }
                    lastRead = unix
                }
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
            // reload maps if necessary
            unix := time.Now().UTC().Unix()
            if  unix-lastRead > interval {
                whitelist := readTxtFile(wlistFile)
                blacklist := readTxtFile(blistFile)
                if verbose > 0 {
                    log.Println("Reload white list:", whitelist)
                    log.Println("Reload black list:", blacklist)
                }
                lastRead = unix
            }
			pat1 := strings.Join(whitelist, "|")
			expect1 := false // match=false means site not in whitelist
			match1, err := regexp.MatchString(pat1, r.URL.Host)
			if err != nil {
				log.Println("ERROR: fail in match", pat1, r.URL.Host)
			}
			pat2 := strings.Join(blacklist, "|")
			expect2 := true // match=true means site is in blacklist
			match2, err := regexp.MatchString(pat2, r.URL.Host)
			if err != nil {
				log.Println("ERROR: fail in match", pat2, r.URL.Host)
			}
			if match2 == expect2 || match1 == expect1 {
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
