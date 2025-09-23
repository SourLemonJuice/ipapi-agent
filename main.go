package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"github.com/SourLemonJuice/ipapi-agent/buildinfo"
	"github.com/SourLemonJuice/ipapi-agent/datasource"
	"github.com/SourLemonJuice/ipapi-agent/resps"
)

var conf config
var queryCache *cache.Cache

func init() {
	log.SetPrefix("ipapi-agent: ")
	log.SetFlags(0)

	queryCache = cache.New(6*time.Hour, 30*time.Minute)
}

func main() {
	var err error

	flag.BoolFunc("version", "print version information of ipapi-agent", flagVersion)
	confPath := flag.String("config", "", "set config file path")
	flag.Parse()

	log.Print("initializing...")

	conf = newConfig()
	// if the path is empty, only default value will be applied.
	if len(*confPath) == 0 {
		log.Print("no config file provided")
	} else {
		log.Printf("loading config file %v", *confPath)
		err = conf.decodeFile(*confPath)
		if err != nil {
			log.Fatalf("can't load config file: %v", err)
		}
	}

	if conf.Dev.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RemoveExtraSlash = true
	router.Use(gin.Recovery())
	if conf.Dev.Log {
		router.Use(gin.Logger())
	}

	router.SetTrustedProxies(conf.TrustedProxies)

	router.GET("/", getRoot)
	router.GET("/query", getQuery)
	router.GET("/query/:addr", getQuery)

	serverAddr := net.JoinHostPort(conf.Listen, strconv.FormatUint(uint64(conf.ListenPort), 10))
	log.Printf("starting server on %v", serverAddr)
	err = router.Run(serverAddr)
	if err != nil {
		log.Fatalf("server(gin) error: %v", err)
	}
}

func flagVersion(s string) error {
	fmt.Printf("ipapi-agent version %v\n\n", buildinfo.Version)

	fmt.Printf("Environment: %v %v/%v\n", buildinfo.GoVersion, buildinfo.OS, buildinfo.Arch)

	os.Exit(0)
	return nil
}

func getRoot(c *gin.Context) {
	// Use \r\n (CRLF) as line break symbol in this function, which it is Windows and HTTP format.
	// Also, don't forget the last line break at the body end.
	var err error

	query := c.ClientIP()
	addrStr, addrIP, err := queryToAddr(query)
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, "[FAILURE]\r\nBad query IP address/domain\r\n")
		return
	}

	if isSpecialAddr(addrIP) {
		c.Abort()
		c.String(http.StatusBadRequest, "[FAILURE]\r\nIP address/domain is in invalid range\r\n")
		return
	}

	var resp resps.Query
	if val, found := queryCache.Get(addrStr); found {
		resp = val.(resps.Query)
		c.String(http.StatusOK, respTXT(addrStr, resp))
		return
	}

	resp.Status = "success"

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addrStr)
	if err != nil {
		log.Printf("Data source error: %v", err)
		c.Abort()
		c.String(http.StatusInternalServerError, "[FAILURE]\r\nData source error: %w\r\n", err)
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		c.Abort()
		c.String(http.StatusInternalServerError, "[FAILURE]\r\nInternal Server Error\r\n")
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.String(http.StatusOK, respTXT(addrStr, resp))
}

func respTXT(ipStr string, resp resps.Query) string {
	var txt strings.Builder

	// U+25CF Black Circle: ●
	txt.WriteString(fmt.Sprintf("\u25cf %v | %v\r\n", ipStr, resp.DataSource))

	tab := tabwriter.NewWriter(&txt, 0, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tab, "Location: \t%v, %v (%v)\r\n", resp.Region, resp.Country, resp.CountryCode)
	fmt.Fprintf(tab, "Timezone: \t%v %v\r\n", resp.Timezone, utcMinToISO8601(resp.UTCOffset))
	fmt.Fprintf(tab, "ISP: \t%v\r\n", resp.ISP)
	fmt.Fprintf(tab, "Org: \t%v\r\n", resp.Org)
	fmt.Fprintf(tab, "ASN: \t%v\r\n", resp.ASN)
	tab.Flush()

	return txt.String()
}

func getQuery(c *gin.Context) {
	var err error

	query := c.Param("addr")
	if query == "" {
		query = c.ClientIP()
	}

	addrStr, addrIP, err := queryToAddr(query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "failure",
			"message": "Bad query IP address/domain: " + err.Error(),
		})
		return
	}

	if isSpecialAddr(addrIP) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "failure",
			"message": "IP address/domain is in invalid range",
		})
		return
	}

	useCache, err := strconv.ParseBool(c.DefaultQuery("cache", "true"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var resp resps.Query
	// love cache ^_^
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(resps.Query)
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.Status = "success"

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addrStr)
	if err != nil {
		log.Printf("Data source error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "failure",
			"message": "Data source error: " + err.Error(),
		})
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		// for security reasons, it won't response error string
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "failure",
			"message": "Internal Server Error",
		})
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.JSON(http.StatusOK, resp)
}

// Convert query string that can contain IP address and domain into one safe IP address format.
// Result won't be: empty string, invalid IP, unresolvable domain.
func queryToAddr(query string) (string, netip.Addr, error) {
	var err error

	if query == "" {
		return "", netip.Addr{}, errors.New("empty IP/domain")
	}

	// query is a real IP address
	addr, err := netip.ParseAddr(query)
	if err == nil {
		return query, addr, nil
	}

	// should we continue parsing?
	if !conf.ResolveDomain {
		return "", netip.Addr{}, errors.New("not permitted to resolve domain")
	}

	// query is a domain name
	addr, err = resolveDomain(query)
	if err != nil {
		return "", netip.Addr{}, err
	}
	return query, addr, nil
}

func resolveDomain(domain string) (netip.Addr, error) {
	if !strings.Contains(domain, ".") {
		return netip.Addr{}, errors.New("invalid domain")
	}
	// block some reserved TLDs
	// you may want to block .lan TLD with config file, because that's not a part of any standard
	blockedTLD := []string{".alt", ".arpa", ".invalid", ".local", ".localhost", ".onion", ".test", ".internal"}
	blockedTLD = append(blockedTLD, conf.Dev.TLDBlockList...)
	for _, tld := range blockedTLD {
		if strings.HasSuffix(domain, tld) {
			return netip.Addr{}, errors.New("invalid domain")
		}
	}

	addrStrArr, err := net.LookupHost(domain)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("lookup domain failure: %w", err)
	}
	addr, err := netip.ParseAddr(addrStrArr[0])
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid domain IP address: %w", err)
	}

	return addr, nil
}

// Check if the given IP is one of loopback, private, unspecified(0.0.0.0), or any non-global unicast address.
func isSpecialAddr(addr netip.Addr) bool {
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsUnspecified() || !addr.IsGlobalUnicast() {
		return true
	}
	return false
}

func utcMinToISO8601(min int) string {
	var out strings.Builder

	out.WriteString("UTC")
	switch {
	case min == 0:
		out.WriteString("0")
		return out.String()
	case min > 0:
		out.WriteString("+")
	case min < 0:
		out.WriteString("-")
		min = -min // our AbsInt() :]
	}

	out.WriteString(fmt.Sprintf("%02d", min/60))
	out.WriteString(fmt.Sprintf("%02d", min%60))
	return out.String()
}
