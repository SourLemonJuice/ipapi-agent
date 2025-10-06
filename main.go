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

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"github.com/SourLemonJuice/ipapi-agent/internal/build"
	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/debug"
	"github.com/SourLemonJuice/ipapi-agent/internal/response"
	"github.com/SourLemonJuice/ipapi-agent/internal/upstream"
)

var (
	conf       config.Config
	queryCache *cache.Cache = cache.New(6*time.Hour, 30*time.Minute)
)

func init() {
	log.SetPrefix("[main] ")
	log.SetFlags(0)

	// force output color, ignore the TTY detection, please
	color.NoColor = false
}

func main() {
	var err error

	flag.BoolFunc("version", "print version information of ipapi-agent", flagVersion)
	confPath := flag.String("config", "", "set config file path")
	flag.Parse()

	log.Print("initializing...")

	err = loadConfig(&conf, *confPath)
	if err != nil {
		log.Fatalln(err)
	}

	if conf.Dev.Debug {
		debug.Enable()
		debug.PrintIntro()
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := upstream.InitSelector(conf.Upstream); err != nil {
		log.Fatalf("failed to initialize upstream selector: %v", err)
	}

	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RemoveExtraSlash = true
	router.Use(gin.Recovery())
	if conf.Dev.Log {
		router.Use(gin.Logger())
	}

	if err := router.SetTrustedProxies(conf.TrustedProxies); err != nil {
		log.Fatalf("invalid trusted proxies configuration: %v", err)
	}

	router.GET("/", getRoot)
	router.GET("/query", getQuery)
	router.GET("/query/:addr", getQuery)

	serverAddr := net.JoinHostPort(conf.Listen, strconv.FormatUint(uint64(conf.Port), 10))
	log.Printf("starting server on %v", serverAddr)
	err = router.Run(serverAddr)
	if err != nil {
		log.Fatalf("server(GIN) error: %v", err)
	}
}

func flagVersion(s string) error {
	build.PrintVersion()
	os.Exit(0)
	return nil
}

func loadConfig(conf *config.Config, hint string) error {
	var err error

	*conf = config.New()

	var path string
	// if no hint use default path
	if len(hint) == 0 {
		confInfo, err := os.Stat("ipapi.toml")
		if err == nil && !confInfo.IsDir() {
			path = "ipapi.toml"
			log.Printf("found config file in default path %v", path)
		}
	} else {
		path = hint
	}

	// if no any file found, only default value will be applied.
	if len(path) != 0 {
		log.Printf("loading config file %v", path)
		err = conf.DecodeFile(path)
		if err != nil {
			return fmt.Errorf("can't load config file: %w", err)
		}
	} else {
		log.Print("no config file provided, use defaults")
	}

	return nil
}

func getRoot(c *gin.Context) {
	// Use \r\n (CRLF) as line break symbol in this function, which it is Windows and HTTP format.
	// Also, don't forget the last line break at the body end.
	var err error

	colorful := false
	if strings.HasPrefix(c.GetHeader("User-Agent"), "curl") {
		colorful = true
	}

	query := c.ClientIP()
	addrStr, addrIP, err := queryToAddr(query)
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, respTXTFailure(colorful, "Bad query IP address/domain"))
		return
	}

	if isSpecialAddr(addrIP) {
		c.Abort()
		c.String(http.StatusBadRequest, respTXTFailure(colorful, "IP address/domain is in invalid range"))
		return
	}

	var resp response.Query
	if val, found := queryCache.Get(addrStr); found {
		resp = val.(response.Query)
		c.String(http.StatusOK, respTXT(colorful, addrStr, resp))
		return
	}

	// let struct cache compatible with getQuery()
	resp.Status = "success"

	apidata, err := upstream.SelectAPI(conf.Upstream)
	if err != nil {
		log.Printf("Upstream selection error: %v", err)
		c.Abort()
		c.String(http.StatusInternalServerError, respTXTFailure(colorful, "Internal Server Error"))
		return
	}
	err = apidata.Request(addrStr)
	if err != nil {
		log.Printf("Data source error: %v", err)
		c.Abort()
		c.String(http.StatusInternalServerError, respTXTFailure(colorful, "Data source error: %v", err))
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		c.Abort()
		c.String(http.StatusInternalServerError, respTXTFailure(colorful, "Internal Server Error"))
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.String(http.StatusOK, respTXT(colorful, addrStr, resp))
}

func respTXTFailure(colorful bool, format string, obj ...any) string {
	var txt strings.Builder
	cRed := color.New(color.FgHiRed)
	if !colorful {
		cRed.DisableColor()
	}

	// U+00D7 Multiplication Sign: ×
	txt.WriteString(cRed.Sprint("\u00d7 FAILURE"))
	txt.WriteString("\r\n")
	txt.WriteString(fmt.Sprintf(format, obj...))
	txt.WriteString("\r\n")

	return txt.String()
}

func respTXT(colorful bool, addrStr string, resp response.Query) string {
	var txt strings.Builder
	cGreen := color.New(color.FgHiGreen)
	if !colorful {
		cGreen.DisableColor()
	}

	// U+25CF Black Circle: ●
	// from systemctl status ^_^
	txt.WriteString(cGreen.Sprint("\u25cf"))
	txt.WriteString(fmt.Sprintf(" %v - %v\r\n", addrStr, resp.DataSource))

	tab := tabwriter.NewWriter(&txt, 0, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tab, "Location: \t%v, %v (%v)\r\n", resp.Region, resp.Country, resp.CountryCode)
	fmt.Fprintf(tab, "Timezone: \t%v %v\r\n", resp.Timezone, utcOffsetToISO8601(resp.UTCOffset))

	if len(resp.Org) == 0 {
		fmt.Fprintf(tab, "Org: \t<Unavailable>\r\n")
	} else {
		fmt.Fprintf(tab, "Org: \t%v\r\n", resp.Org)
	}
	if len(resp.ISP) > 0 {
		fmt.Fprintf(tab, "ISP: \t%v\r\n", resp.ISP)
	}

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

	var resp response.Query
	// love cache ^_^
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(response.Query)
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.Status = "success"

	apidata, err := upstream.SelectAPI(conf.Upstream)
	if err != nil {
		log.Printf("Upstream selection error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "failure",
			"message": "Internal Server Error",
		})
		return
	}
	err = apidata.Request(addrStr)
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
	if !conf.Resolve.Domain {
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
	// you may want to block .lan TLD with config file, because that's not a part of any standard.
	// https://en.wikipedia.org/wiki/Special-use_domain_name
	blockedTLD := []string{".alt", ".arpa", ".invalid", ".local", ".localhost", ".onion", ".test", ".internal"}
	blockedTLD = append(blockedTLD, conf.Resolve.BlockTLD...)
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

func utcOffsetToISO8601(min int) string {
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
