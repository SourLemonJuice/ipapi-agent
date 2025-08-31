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

	var isIP bool
	if net.ParseIP(addr) != nil {
		isIP = true
	} else {
		isIP = false
	}

	var resp respstruct.Query
	if val, found := queryCache.Get(addr); found {
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

	if isIP {
		queryCache.Set(addr, resp, 4*time.Hour)
	} else {
		queryCache.Set(addr, resp, 2*time.Minute)
	}

	c.JSON(http.StatusOK, resp)
}
