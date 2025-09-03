package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
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
	// Please use \r\n (CRLF) as line break symbol in this function, which it is Windows and HTTP format.
	// Also, don't forget the last line break at the body end.
	var err error

	// query := c.ClientIP() // TODO
	query := "1.1.1.1"
	addrStr, addrIP, err := addrToIP(query)
	if err != nil {
		c.Abort()
		c.String(http.StatusOK, "FAILURE\r\nCan not parse you IP address\r\n")
		return
	}

	// TODO restore it
	// if isSpecialIP(addrIP) {
	// 	c.Abort()
	// 	c.String(http.StatusOK, "FAILURE\r\nYou IP address is special\r\n")
	// 	return
	// }
	fmt.Println(addrStr, addrIP)

	useCache, err := strconv.ParseBool(c.DefaultQuery("cache", "true"))
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, "HTTP 400 Bad Request\r\n")
		return
	}

	var resp respstruct.Query
	if val, found := queryCache.Get(addrStr); found && useCache {
		resp = val.(respstruct.Query)
		c.String(http.StatusOK, respTXT(resp))
		return
	}

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addrStr)
	if err != nil {
		c.Abort()
		c.String(http.StatusOK, "FAILURE\r\nData source requesting error: %w\r\n", err)
		return
	}

	if apidata.IsSuccess() {
		resp.Status = "success"
	} else {
		c.Abort()
		c.String(http.StatusOK, "FAILURE\r\nData source response error: %v\r\n", apidata.GetMessage())
		return
	}

	err = apidata.Fill(&resp)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	queryCache.SetDefault(addrStr, resp)

	c.String(http.StatusOK, respTXT(resp))
}

func respTXT(resp respstruct.Query) string {
	var txt string

	txt += strings.ToUpper(resp.Status) + "\r\n"
	txt += "Data Source:\t" + resp.DataSource + "\r\n"
	txt += "Country:\t" + resp.Country + "\r\n"
	txt += "Country Code:\t" + resp.CountryCode + "\r\n"
	txt += "Region:\t\t" + resp.Region + "\r\n"
	txt += "Timezone:\t" + resp.Timezone + "\r\n"
	txt += "UTC Offset(min):\t" + strconv.FormatInt(int64(resp.UTCOffset), 10) + "\r\n"
	txt += "ISP:\t\t" + resp.ISP + "\r\n"
	txt += "Organization:\t" + resp.Org + "\r\n"
	txt += "ASN:\t\t" + resp.ASN + "\r\n"

	return txt
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
			"message": "Query IP address/domain is special",
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
			"message": "Data source requesting error: " + err.Error(),
		})
		return
	}

	if apidata.IsSuccess() {
		resp.Status = "success"
	} else {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Data source response error: " + apidata.GetMessage(),
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
