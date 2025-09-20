package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"github.com/SourLemonJuice/ipapi-agent/datasource"
	"github.com/SourLemonJuice/ipapi-agent/respstruct"
)

var queryCache *cache.Cache

func main() {
	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RemoveExtraSlash = true
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.GET("/", getRoot)
	router.GET("/query", getQuery)
	router.GET("/query/:addr", getQuery)

	queryCache = cache.New(6*time.Hour, 30*time.Minute)

	router.Run(":8080")
}

func getRoot(c *gin.Context) {
	// Use \r\n (CRLF) as line break symbol in this function, which it is Windows and HTTP format.
	// Also, don't forget the last line break at the body end.
	var err error

	query := c.ClientIP()
	addrStr, addrIP, err := addrToIP(query)
	if err != nil {
		c.Abort()
		c.String(http.StatusOK, "[FAILURE]\r\nBad query IP address/domain\r\n")
		return
	}

	if isSpecialIP(addrIP) {
		c.Abort()
		c.String(http.StatusOK, "[FAILURE]\r\nIP address/domain is in invalid range\r\n")
		return
	}

	useCache, err := strconv.ParseBool(c.DefaultQuery("cache", "true"))
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, "HTTP 400 Bad Request\r\n") // display for human
		return
	}

	var resp respstruct.Query
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(respstruct.Query)
		c.String(http.StatusOK, respTXT(addrStr, resp))
		return
	}

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addrStr)
	if err != nil {
		c.Abort()
		c.String(http.StatusOK, "[FAILURE]\r\nRequest failure: %w\r\n", err)
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.String(http.StatusOK, respTXT(addrStr, resp))
}

func respTXT(ipStr string, resp respstruct.Query) string {
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

	addrStr, addrIP, err := addrToIP(query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Bad query IP address/domain: " + err.Error(),
		})
		return
	}

	if isSpecialIP(addrIP) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
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

	var resp respstruct.Query
	// love cache ^_^
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(respstruct.Query)
		c.JSON(http.StatusOK, resp)
		return
	}

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addrStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Request failure: " + err.Error(),
		})
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.JSON(http.StatusOK, resp)
}

// Convert query string that can contain IP address and domain into one safe IP address format.
// Result won't be: empty string, invalid IP, unresolvable domain.
func addrToIP(query string) (string, net.IP, error) {
	if query == "" {
		// string can be nil
		return "", nil, errors.New("empty IP/domain")
	}

	// query is a real IP address
	ip := net.ParseIP(query)
	if ip != nil {
		return query, ip, nil
	}

	// query is a domain name
	ipStrArr, err := net.LookupHost(query)
	if err != nil {
		return "", nil, fmt.Errorf("lookup addr as domain failure: %w", err)
	}
	ip = net.ParseIP(ipStrArr[0])
	if ip == nil {
		return "", nil, errors.New("invalid domain IP address")
	}
	return ipStrArr[0], ip, nil
}

// Check if the given IP is one of loopback, private, unspecified(0.0.0.0), or any non-global unicast address.
func isSpecialIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || !ip.IsGlobalUnicast() {
		return true
	}
	return false
}

func utcMinToISO8601(min int) string {
	var out strings.Builder

	out.WriteString("UTC")
	if min == 0 {
		out.WriteString("0")
		return out.String()
	} else if min > 0 {
		out.WriteString("+")
	} else if min < 0 {
		out.WriteString("-")
		min = -min // our AbsInt() :]
	}

	out.WriteString(fmt.Sprintf("%02d", min/60))
	out.WriteString(fmt.Sprintf("%02d", min%60))
	return out.String()
}
