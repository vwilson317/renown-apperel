package main

import (
	// "encoding/json"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type FindingReponse struct {
	FindItemsIneBayStoresResponse []struct {
		Ack          []string    `json:"ack"`
		Version      []string    `json:"version"`
		Timestamp    []time.Time `json:"timestamp"`
		SearchResult []struct {
			Count string `json:"@count"`
			Item  []struct {
				ItemID          []string `json:"itemId"`
				Title           []string `json:"title"`
				GlobalID        []string `json:"globalId"`
				PrimaryCategory []struct {
					CategoryID   []string `json:"categoryId"`
					CategoryName []string `json:"categoryName"`
				} `json:"primaryCategory"`
				GalleryURL    []string `json:"galleryURL"`
				ViewItemURL   []string `json:"viewItemURL"`
				PaymentMethod []string `json:"paymentMethod"`
				AutoPay       []string `json:"autoPay"`
				PostalCode    []string `json:"postalCode"`
				Location      []string `json:"location"`
				Country       []string `json:"country"`
				ShippingInfo  []struct {
					ShippingServiceCost []struct {
						CurrencyID string `json:"@currencyId"`
						Value      string `json:"__value__"`
					} `json:"shippingServiceCost"`
					ShippingType            []string `json:"shippingType"`
					ShipToLocations         []string `json:"shipToLocations"`
					ExpeditedShipping       []string `json:"expeditedShipping"`
					OneDayShippingAvailable []string `json:"oneDayShippingAvailable"`
					HandlingTime            []string `json:"handlingTime"`
				} `json:"shippingInfo"`
				SellingStatus []struct {
					CurrentPrice []struct {
						CurrencyID string `json:"@currencyId"`
						Value      string `json:"__value__"`
					} `json:"currentPrice"`
					ConvertedCurrentPrice []struct {
						CurrencyID string `json:"@currencyId"`
						Value      string `json:"__value__"`
					} `json:"convertedCurrentPrice"`
					SellingState []string `json:"sellingState"`
					TimeLeft     []string `json:"timeLeft"`
				} `json:"sellingStatus"`
				ListingInfo []struct {
					BestOfferEnabled  []string    `json:"bestOfferEnabled"`
					BuyItNowAvailable []string    `json:"buyItNowAvailable"`
					StartTime         []time.Time `json:"startTime"`
					EndTime           []time.Time `json:"endTime"`
					ListingType       []string    `json:"listingType"`
					Gift              []string    `json:"gift"`
					WatchCount        []string    `json:"watchCount"`
				} `json:"listingInfo"`
				ReturnsAccepted []string `json:"returnsAccepted"`
				Condition       []struct {
					ConditionID          []string `json:"conditionId"`
					ConditionDisplayName []string `json:"conditionDisplayName"`
				} `json:"condition"`
				IsMultiVariationListing []string `json:"isMultiVariationListing"`
				TopRatedListing         []string `json:"topRatedListing"`
			} `json:"item"`
		} `json:"searchResult"`
		PaginationOutput []struct {
			PageNumber     []string `json:"pageNumber"`
			EntriesPerPage []string `json:"entriesPerPage"`
			TotalPages     []string `json:"totalPages"`
			TotalEntries   []string `json:"totalEntries"`
		} `json:"paginationOutput"`
		ItemSearchURL []string `json:"itemSearchURL"`
	} `json:"findItemsIneBayStoresResponse"`
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

func findingHandler(c *gin.Context) {
	pageNum := c.DefaultQuery("pageNum", "1")
	keywords := c.DefaultQuery("keywords", "")

	var responseData = getData(keywords, 20, pageNum)
	c.String(http.StatusOK, string(responseData))
}

func healthHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func listingPostHandler(c *gin.Context) {
	key := c.DefaultQuery("itemId", "")
	isInCart := c.DefaultQuery("isInCart", "true")
	parsed, _ := strconv.ParseBool(isInCart)
	err := client.Set(key, parsed, time.Hour).Err()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		c.Status(http.StatusCreated)
	}
}

func listingGetHandler(c *gin.Context) {
	key := c.DefaultQuery("itemId", "")
	value, err := client.Get(key).Result()
	if err != nil && err.Error() != "redis: nil" {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		var boolVal = value == "1"
		c.String(http.StatusOK, strconv.FormatBool(boolVal))
	}
}

func getData(keywords string, pageSize int, pageNum string) []byte {
	pageSizeStr := strconv.Itoa(pageSize)
	var queryStr = "https://svcs.ebay.com/services/search/FindingService/v1?OPERATION-NAME=findItemsIneBayStores&SERVICE-VERSION=1.13.0&SECURITY-APPNAME=VincentW-renownap-PRD-0b31f104d-07a63429&RESPONSE-DATA-FORMAT=JSON&REST-PAYLOAD=&paginationInput.entriesPerPage=" + pageSizeStr + "&paginationInput.pageNumber=" + pageNum + "&storeName=imyown"
	queryStr += "&keywords=" + keywords
	var data []byte
	response, err := http.Get(queryStr)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ = ioutil.ReadAll(response.Body)
	}
	return data
}

func popHandler(c *gin.Context) {
	pageSize := 100
	var (
		// b             []byte
		stringBuilder strings.Builder
		// jsonData      string
		responseObj FindingReponse
		totalPages  int = 2 // assuming there's at least 2 pages
	)

	for i := 1; i < totalPages; i++ {
		pageNum := strconv.Itoa(i)
		var responseByteArray = getData("", pageSize, pageNum)
		stringBuilder.Grow(len(responseByteArray))
		stringBuilder.Write(responseByteArray)

		err := json.Unmarshal(responseByteArray, &responseObj)
		if err != nil {
			fmt.Println("error:", err)
		}
		// c.String(http.StatusOK, stringBuilder.String())
		totalPages, _ = strconv.Atoi(responseObj.FindItemsIneBayStoresResponse[0].PaginationOutput[0].TotalPages[0])
		addToEs(responseObj, i, c)
	}

	// c.String(http.StatusOK, "")
}

func addToEs(response FindingReponse, pageNum int, c *gin.Context) {
	res, err := esClient.Indices.Create("listings")
	if err != nil {
		fmt.Printf("Cannot create index: %s", err)
	}
	if res.IsError() {
		fmt.Printf("Cannot create index: %s", res)
	}
	var (
	// strBuilder strings.Builder
	// raw       map[string]interface{}
	// blk       *bulkResponse
	// numErrors int
	// buffer bytes.Buffer
	)

	items := response.FindItemsIneBayStoresResponse[0].SearchResult[0].Item

	if err != nil {
		log.Fatalf("Error creating the indexer: %s", err)
	}

	for _, rec := range items {
		itemID := rec.ItemID[0]
		data, _ := json.Marshal(rec)

		req := esapi.IndexRequest{
			Index:      "listings",
			DocumentID: itemID,
			Body:       bytes.NewReader(data),
			Refresh:    "true",
		}

		res, err = req.Do(context.Background(), esClient)
		if err != nil {
			fmt.Printf("Error getting response: %s %s", err, "/n")
		}
		res.Body.Close()

	}
}

func initClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:8086",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}

func initEsClient() *elasticsearch.Client {
	// cfg := elasticsearch.Config{
	// 	Addresses: []string{"http://localhost:9200", "http://localhost:9300"},
	// 	Transport: &http.Transport{
	// 		MaxIdleConnsPerHost: 10,
	// 		// ResponseHeaderTimeout: time.Millisecond,
	// 		// DialContext:           (&net.Dialer{Timeout: time.Nanosecond}).DialContext,
	// 		TLSClientConfig: &tls.Config{
	// 			MinVersion: tls.VersionTLS11,
	// 			// ...
	// 		},
	// 	},
	// }

	es, err := elasticsearch.NewDefaultClient() //elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("Error creating the es client: %s", err)
	}

	return es
}

var client *redis.Client
var esClient *elasticsearch.Client

func main() {
	fmt.Println("Starting the application...")

	router := gin.Default()

	client = initClient()
	esClient = initEsClient()

	router.GET("/api/finding", findingHandler)
	router.GET("/health", healthHandler)
	router.GET("/api/populate", popHandler)

	router.POST("/api/listings", listingPostHandler)
	router.GET("/api/listings/cartstatus", listingGetHandler)

	router.Use(cors.Default())

	router.Run(":8083")

	fmt.Println("Terminating the application...")
}
