package main

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"github.com/SourLemonJuice/ipapi-agent/datasource"
	"github.com/SourLemonJuice/ipapi-agent/respstruct"
)

var queryCache *cache.Cache

const (
	ipExpiration     = 4 * time.Hour
	domainExpiration = 2 * time.Minute
)

func main() {
	router := gin.Default()

	router.GET("/query/:addr", getQuery)

	// the first expiration time at here is just a fallback
	queryCache = cache.New(1*time.Hour, 30*time.Minute)

	router.Run(":8080")
}

func getQuery(c *gin.Context) {
	var err error

	addr := c.Param("addr")
	if addr == "" {
		c.AbortWithStatus(http.StatusNotFound)
	}

	var addrIsIP bool
	if net.ParseIP(addr) != nil {
		addrIsIP = true
	} else {
		addrIsIP = false
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
	if val, found := queryCache.Get(addr); found && useCache {
		resp = val.(respstruct.Query)
		// love cache ^_^
		c.JSON(http.StatusOK, resp)
		return
	}

	var apidata datasource.Interface = &datasource.IpapiCom{}
	err = apidata.DoRequest(addr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Data source requesting error",
		})
		return
	}

	if !apidata.IsSuccess() {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  "failure",
			"message": "Data source response error: " + apidata.GetMessage(),
		})
		return
	} else {
		resp.Status = "success"
	}

	err = apidata.Fill(&resp)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if addrIsIP {
		queryCache.Set(addr, resp, ipExpiration)
	} else {
		queryCache.Set(addr, resp, domainExpiration)
	}

	c.JSON(http.StatusOK, resp)
}
