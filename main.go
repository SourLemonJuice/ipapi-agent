package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
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

	router.GET("/query", getQuery)
	router.GET("/query/:addr", getQuery)

	queryCache = cache.New(6*time.Hour, 30*time.Minute)

	router.Run(":8080")
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
	}

	if addrIP.IsLoopback() {
		c.Abort()
		c.String(http.StatusTeapot, "You are the Creator? You come from a familiar space.")
		return
	}
	if isSpecialIP(addrIP) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Query IP address/domain is special",
		})
		return
	}

	var useCache bool
	switch c.DefaultQuery("cache", "true") {
	case "true":
		useCache = true
	case "false":
		useCache = false
	default:
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

	queryCache.Set(addrStr, resp, cache.DefaultExpiration)

	c.JSON(http.StatusOK, resp)
}

// Convert query string that can contain IP address and domain into one safe IP address format.
// Result won't be: empty string, invalid IP, unresolvable domain.
func addrToIP(addr string) (string, net.IP, error) {
	if addr == "" {
		// string can be nil
		return "", nil, errors.New("Empty IP/domain")
	}

	ip := net.ParseIP(addr)
	if ip != nil {
		return addr, ip, nil
	}

	ips, err := net.LookupHost(addr)
	if err != nil {
		return "", nil, fmt.Errorf("Lookup addr as domain failure: %w", err)
	}
	ip = net.ParseIP(ips[0])
	if ip == nil {
		return "", nil, errors.New("Invalid domain IP address")
	}
	return ips[0], ip, nil
}

// Check if the given IP is one of loopback, private, unspecified(0.0.0.0), or not global unicast address.
func isSpecialIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || !ip.IsGlobalUnicast() {
		return true
	}
	return false
}
