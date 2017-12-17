package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	serviceURL  string
	servicePort int
	secret      string
}

type EFApi struct {
	config            *Config
	productController *ProductsController
}

func CreateEFSApi(config *Config) *EFApi {
	efApi := &EFApi{
		config:            config,
		productController: &ProductsController{},
	}

	efApi.productController.config = config

	efApi.productController.buildDataModel()

	return efApi
}

func CreateNewEFSRouter(efApi *EFApi) *httprouter.Router {
	router := httprouter.New()

	router.POST("/list_categories", efApi.productController.listAllCategories)
	router.POST("/list_products_by", efApi.productController.listAllProductsByCategory)

	router.POST("/show_product", efApi.productController.showProductDetails)

	router.GET("/rebuild_data/:secret_key", efApi.productController.rebuildDataModel)

	router.ServeFiles("/resources/*filepath", http.Dir("./resources"))

	return router
}

func main() {
	configContent, err := ioutil.ReadFile("config.yml")
	configParams := make(map[string]string)

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal([]byte(configContent), &configParams)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	port := configParams["port"]

	if port == "" {
		log.Fatal("Error with port number configuration")
	}

	serverPortI, _ := strconv.Atoi(port)

	url := configParams["url"]

	secretKey := configParams["secret"]

	if secretKey == "" {
		log.Fatal("Error with port number configuration")
	}

	config := &Config{
		serviceURL:  url,
		servicePort: serverPortI,
		secret:      secretKey,
	}

	efsApi := CreateEFSApi(config)
	router := CreateNewEFSRouter(efsApi)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
