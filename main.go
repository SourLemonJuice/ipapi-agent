package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SourLemonJuice/ipapi-agent/datasource"
	"github.com/SourLemonJuice/ipapi-agent/respstruct"
)

func main() {
	router := gin.Default()

	router.GET("/query/:addr", getQuery)

	router.Run(":8080")
}

func getQuery(c *gin.Context) {
	var err error

	addr := c.Param("addr")
	if addr == "" {
		c.AbortWithStatus(http.StatusNotFound)
	}

	var resp respstruct.Query
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

	c.JSON(http.StatusOK, resp)
}
