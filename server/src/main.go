package main

import (
	// "encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func findingHandler(c *gin.Context) {
	pageNum := c.DefaultQuery("pageNum", "1")
	response, err := http.Get("https://svcs.ebay.com/services/search/FindingService/v1?OPERATION-NAME=findItemsIneBayStores&SERVICE-VERSION=1.13.0&SECURITY-APPNAME=VincentW-renownap-PRD-0b31f104d-07a63429&RESPONSE-DATA-FORMAT=JSON&REST-PAYLOAD=&paginationInput.entriesPerPage=20&paginationInput.pageNumber=" + pageNum + "&storeName=imyown")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		c.String(http.StatusOK, string(data))
	}
}

func healthHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func main() {
	fmt.Println("Starting the application...")

	router := gin.Default()

	router.GET("/api/finding", findingHandler)
	router.GET("/health", healthHandler)

	router.Use(cors.Default())
	router.Run(":8083")

	fmt.Println("Terminating the application...")
}