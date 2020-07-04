package cc

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var Config Cfg

// proxyUrl add proxy Config to help DefaultTransport send the request to
// urls specified in the configuration
type proxyUrl struct {
	Scheme string         `toml:"Scheme"`
	Host   string         `toml:"Host"`
	Path   string         `toml:"Path"`
	U      *url.URL       `toml:"-"`
	Re     *regexp.Regexp `toml:"-"`
}

type SubFilter struct {
	Old         string         `toml:"Old"`
	Repl        string         `toml:"Repl"`
	Path        string         `toml:"Path"`
	RUriMatcher *regexp.Regexp `toml:"-"`
	OldMatcher  *regexp.Regexp
}

func (s *SubFilter) HandleLine(line []byte) []byte {
	return s.OldMatcher.ReplaceAll(line, []byte(s.Repl))
}

// CustomHeader help resp
type CustomHeader struct {
	Key   string `toml:"Key"`
	Value string `toml:"Value"`
}

type Cfg struct {
	ListenPort    int    `toml:"ListenPort"`
	DefaultTarget string `toml:"DefaultTarget"`
	PPROFPort     int    `toml:"PPROFPort"`
	LogFlag       string `toml:"LogFlag"`
	Middleware    string `toml:"Middleware"`
	Mode          string `toml:"Mode"`
	// like proxy_pass
	ProxyUrls map[string]proxyUrl `toml:"ProxyUrls"`

	// add custom header
	CustomHeaders map[string]CustomHeader `toml:"CustomHeaders"`

	// sub_filter
	SubFilters map[string]SubFilter `toml:"SubFilter"`

	// js lua scripts
	LuaPath string `toml:"LuaPath"`
	JsPath  string `toml:"JSPath"`
}

// InitConfig pass in a filename and reread all config from file to cover origin value
func (c *Cfg) InitConfig(filename string) error {
	var err error
	log.Printf("load config from %s", filename)
	if _, err = toml.DecodeFile(filename, c); err != nil {
		return err
	}
	c.SetLogFlag()
	// print config after flag was set
	c.Print()
	if err = c.parseProxyUrls(); err != nil {
		return err
	}
	if err = c.parseSubFilters(); err != nil {
		return err
	}
	return nil
}

// Print use reflect package to print all config
func (c *Cfg) Print() {
	key := reflect.TypeOf(*c)
	value := reflect.ValueOf(*c)
	for i := 0; i < value.NumField(); i++ {
		field := key.Field(i)
		log.Printf("key=%v value=%v", field.Name, value.Field(i).Interface())
	}
}

// parseProxyUrls parse string proxy url with regexp package
func (c *Cfg) parseProxyUrls() error {
	for k := range c.ProxyUrls {
		u, err := url.Parse(fmt.Sprintf("%s://%s", c.ProxyUrls[k].Scheme, c.ProxyUrls[k].Host))
		if err != nil {
			return err
		}
		temp := c.ProxyUrls[k]
		temp.U = u
		temp.Re = regexp.MustCompile(c.ProxyUrls[k].Path)
		log.Printf("parseProxyUrl url:[%s], proxy:[%s]", temp.Re, u.RequestURI())
		c.ProxyUrls[k] = temp
	}
	log.Printf("parseProxyUrls ok! len(ProxyUrls)=[%d]", len(c.ProxyUrls))
	return nil
}

func (c *Cfg) parseSubFilters() error {
	var err error
	for sub := range c.SubFilters {
		f := c.SubFilters[sub]
		f.RUriMatcher, err = regexp.Compile(f.Path)
		if err != nil {
			return err
		}
		f.OldMatcher, err = regexp.Compile(f.Old)
		if err != nil {
			return err
		}
		log.Printf("load subFilter on url:[%s], filter:[%s]->[%s]", f.Path, f.Old, f.Repl)
		c.SubFilters[sub] = f
	}
	log.Printf("parseSubFilters ok! len(SubFilter)=[%d]", len(c.SubFilters))
	return nil
}

// SetLogFlag set global config flag
func (c *Cfg) SetLogFlag() {
	logC := strings.Split(c.LogFlag, "")
	logFlag := log.Ltime
	for _, c := range logC {
		switch c {
		case "d":
			logFlag |= log.Ldate
		case "l":
			logFlag |= log.Lshortfile
		default:
			continue
		}
	}
	log.SetFlags(logFlag)
}
