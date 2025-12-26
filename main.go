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
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/publicsuffix"

	"github.com/SourLemonJuice/ipapi-agent/build"
	"github.com/SourLemonJuice/ipapi-agent/config"
	C "github.com/SourLemonJuice/ipapi-agent/constant"
	"github.com/SourLemonJuice/ipapi-agent/debug"
	"github.com/SourLemonJuice/ipapi-agent/response"
	"github.com/SourLemonJuice/ipapi-agent/upstream"
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

	err = findConfig(&conf, *confPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if conf.Dev.Debug {
		debug.Enable()
		debug.PrintIntro()
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	upstream.InitSelector(conf.Upstream)

	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RemoveExtraSlash = true
	router.Use(gin.Recovery())
	if conf.Dev.Log {
		router.Use(gin.Logger())
	}

	err = router.SetTrustedProxies(conf.TrustedProxies)
	if err != nil {
		log.Printf("can't set trusted proxies: %v", err)
		os.Exit(1)
	}

	router.GET("/", getRoot)
	router.GET("/query", getQuery)
	router.GET("/query/:addr", getQuery)

	router.GET("/generate_204", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	serverAddr := net.JoinHostPort(conf.Listen, strconv.FormatUint(uint64(conf.Port), 10))
	log.Printf("starting server on %v", serverAddr)
	err = router.Run(serverAddr)
	if err != nil {
		log.Printf("server(GIN) error: %v", err)
		os.Exit(1)
	}
}

func flagVersion(s string) error {
	build.PrintVersion()
	os.Exit(0)
	return nil
}

func findConfig(conf *config.Config, hint string) error {
	var err error

	*conf = config.Default()

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

func getRoot(ctx *gin.Context) {
	// Use \r\n (CRLF) as line break symbol in this function, which it is Windows and HTTP format.
	// Also, don't forget the last line break at the body end.
	var err error

	colorful := false
	if strings.HasPrefix(ctx.GetHeader("User-Agent"), "curl") {
		colorful = true
	}

	query := ctx.ClientIP()
	addrStr, err := parseQuery(query)
	if err != nil {
		log.Printf("Bad IP address/domain: %v", err)
		ctx.Abort()
		ctx.String(http.StatusBadRequest, respTXTFailure(colorful, "Bad IP address/domain"))
		return
	}

	var resp response.Query
	if val, found := queryCache.Get(addrStr); found {
		resp = val.(response.Query)
		ctx.String(http.StatusOK, respTXT(colorful, addrStr, resp))
		return
	}

	api, err := upstream.SelectAPI(conf.Upstream)
	if err != nil {
		log.Fatalf("Can't select API: %v", err)
		ctx.Abort()
		ctx.String(http.StatusInternalServerError, respTXTFailure(colorful, "Internal Server Error"))
		return
	}

	resp, err = api.Fetch(addrStr)
	if err != nil {
		log.Printf("Upstream error: %v", err)
		ctx.Abort()
		ctx.String(http.StatusInternalServerError, respTXTFailure(colorful, "Upstream error"))
		return
	}

	// let struct cache compatible with getQuery()
	resp.Status = C.ResponseStatusSuccess

	queryCache.SetDefault(addrStr, resp)

	ctx.String(http.StatusOK, respTXT(colorful, addrStr, resp))
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
	cYellow := color.New(color.FgHiYellow, color.Bold)
	if !colorful {
		cGreen.DisableColor()
		cYellow.DisableColor()
	}

	// U+25CF Black Circle: ●
	// from systemctl status ^_^
	txt.WriteString(cGreen.Sprint("\u25cf"))
	txt.WriteString(fmt.Sprintf(" %v", addrStr))
	if resp.Anycast {
		txt.WriteString(cYellow.Sprint(" (Anycast)"))
	}
	txt.WriteString(fmt.Sprintf(" - %v\r\n", resp.DataSource))

	tab := tabwriter.NewWriter(&txt, 2, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tab, "\tLocation: \t%v, %v (%v)\r\n", resp.Region, resp.Country, resp.CountryCode)
	fmt.Fprintf(tab, "\tTimezone: \t%v %v\r\n", resp.Timezone, utcOffsetToISO8601(resp.UTCOffset))

	if len(resp.Org) == 0 {
		fmt.Fprintf(tab, "\tOrg: \t<Unavailable>\r\n")
	} else {
		fmt.Fprintf(tab, "\tOrg: \t%v\r\n", resp.Org)
	}
	if len(resp.ISP) > 0 {
		fmt.Fprintf(tab, "\tISP: \t%v\r\n", resp.ISP)
	}

	fmt.Fprintf(tab, "\tASN: \t%v\r\n", resp.ASN)
	tab.Flush()

	return txt.String()
}

func getQuery(ctx *gin.Context) {
	var err error

	query := ctx.Param("addr")
	if query == "" {
		query = ctx.ClientIP()
	}

	addrStr, err := parseQuery(query)
	if err != nil {
		log.Printf("Bad IP address/domain: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  C.ResponseStatusFailure,
			"message": "Bad IP address/domain",
		})
		return
	}

	useCache, err := strconv.ParseBool(ctx.DefaultQuery("cache", "true"))
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var resp response.Query
	// love cache ^_^
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(response.Query)
		ctx.JSON(http.StatusOK, resp)
		return
	}

	api, err := upstream.SelectAPI(conf.Upstream)
	if err != nil {
		log.Fatalf("Can't select API: %v", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	resp, err = api.Fetch(addrStr)
	if err != nil {
		log.Printf("Upstream error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  C.ResponseStatusFailure,
			"message": "Upstream error",
		})
		return
	}

	resp.Status = C.ResponseStatusSuccess

	queryCache.SetDefault(addrStr, resp)

	ctx.JSON(http.StatusOK, resp)
}

// Convert query string that can contain IP address and domain into one safe IP address format.
// Result won't be: empty string, invalid IP, unresolvable domain.
func parseQuery(query string) (addrStr string, err error) {
	if query == "" {
		return "", errors.New("empty query")
	}

	// return the query if it's a real IP address
	addrIP, err := netip.ParseAddr(query)
	if err == nil {
		if isSpecialAddr(addrIP) {
			return "", errors.New("IP address is in invalid range")
		}
		return query, nil
	}

	// should we continue parsing?
	if !conf.Domain.Enabled {
		return "", errors.New("not permitted to resolve domain")
	}

	// query is a domain name, resolve it
	addrStr, err = resolveDomain(query)
	if err != nil {
		return addrStr, err
	}

	addrIP, err = netip.ParseAddr(addrStr)
	if err != nil {
		return addrStr, err
	}
	if isSpecialAddr(addrIP) {
		return "", errors.New("IP address is in invalid range")
	}

	return addrStr, nil
}

func resolveDomain(domain string) (addrStr string, err error) {
	// check its suffix
	suffix, _ := publicsuffix.PublicSuffix(domain)
	if slices.Contains(conf.Domain.BlockSuffix, suffix) {
		return "", errors.New("invalid domain suffix")
	}

	addrStrArr, err := net.LookupHost(domain)
	if err != nil {
		return "", fmt.Errorf("lookup domain failure: %w", err)
	}
	addrStr = addrStrArr[0]

	return addrStr, nil
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
