package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	yaml "gopkg.in/yaml.v2"
)

type ProductsController struct {
	config *Config

	categoriesArray []*CategoryModel
	productsArray   map[int][]interface{}

	productsInCategory map[int]int
}

type CategoryModel struct {
	CategoryID   int    `json:"id"`
	CategoryName string `json:"name"`
}

type ProductModel struct {
	ProductID            int             `json:"id"`
	CategoryID           int             `json:"category_id"`
	ProductName          string          `json:"name"`
	ProductDescription   string          `json:"description"`
	ProductCode          int             `json:"code"`
	ProductQantity       int             `json:"qty"`
	ProductPrice         float64         `json:"price"`
	ProductDiscount      bool            `json:"discount"`
	ProductDiscountPrice float64         `json:"discount_price"`
	ProductImages        []*ProductImage `json:"product_images"`
}

type ProductImage struct {
	ImageName string `json:"image_name"`
}

func (productsController *ProductsController) buildDataModel() {
	categories := productsController.buildCategories()
	productsController.categoriesArray = categories

	products := productsController.buildProducts()
	productsController.productsArray = products
}

func (productsController *ProductsController) rebuildDataModel(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	secretKey := params.ByName("secret_key")

	if secretKey == productsController.config.secret {
		productsController.productsArray = nil
		productsController.categoriesArray = nil
		productsController.productsInCategory = nil

		productsController.buildDataModel()

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
}

func (productsController *ProductsController) buildCategories() []*CategoryModel {
	categoriesArray := make([]*CategoryModel, 0)

	content, err := ioutil.ReadFile("./resources/resources.yml")

	categories := make(map[string][]string)

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal([]byte(content), &categories)
	if err != nil {
		log.Fatalf("error: %v", categories)
	}

	listOfCategories := categories["categories"]

	if len(listOfCategories) > 0 {
		for index, val := range listOfCategories {
			model := &CategoryModel{
				CategoryID:   index,
				CategoryName: val,
			}

			categoriesArray = append(categoriesArray, model)
		}
	}

	return categoriesArray
}

func (productsController *ProductsController) buildProducts() map[int][]interface{} {
	productsMap := make(map[int][]interface{})

	categoriesArray := productsController.categoriesArray

	productsIndexCounter := 0

	for _, value := range categoriesArray {
		catContent, err := ioutil.ReadFile("./resources/categories/" + value.CategoryName + "/products.yml")

		products := make(map[string][]string)

		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal([]byte(catContent), &products)
		if err != nil {
			log.Fatalf("error: %v", products)
		}

		productsList := products["products"]

		if len(productsList) > 0 {
			productsController.productsInCategory = make(map[int]int)
			for _, pvalue := range productsList {
				prodContent, err := ioutil.ReadFile("./resources/categories/" + value.CategoryName + "/products/" + pvalue + "/details.yml")

				productDetails := make(map[string]interface{})

				if err != nil {
					panic(err)
				}

				err = yaml.Unmarshal([]byte(prodContent), &productDetails)
				if err != nil {
					log.Fatalf("error: %v", products)
				}

				productModel := productsController.createProductModelFromData(productsIndexCounter, value, productDetails)
				productsMap[value.CategoryID] = append(productsMap[value.CategoryID], productModel)

				productsController.productsInCategory[productsIndexCounter] = value.CategoryID

				productsIndexCounter = productsIndexCounter + 1
			}
		}
	}

	return productsMap
}

func (productsController *ProductsController) createProductModelFromData(productID int, categoryModel *CategoryModel, productDetails map[string]interface{}) *ProductModel {
	productModel := &ProductModel{
		ProductID:            productID,
		CategoryID:           categoryModel.CategoryID,
		ProductName:          productDetails["name"].(string),
		ProductDescription:   productDetails["description"].(string),
		ProductCode:          productDetails["code"].(int),
		ProductQantity:       productDetails["quantity"].(int),
		ProductPrice:         productDetails["price"].(float64),
		ProductDiscount:      productDetails["discount"].(bool),
		ProductDiscountPrice: productDetails["disprice"].(float64),
	}

	imagesArray := productDetails["images"].([]interface{})

	if len(imagesArray) > 0 {
		productImages := make([]*ProductImage, 0)

		for _, value := range imagesArray {
			imageName := value.(string)
			imageModel := &ProductImage{
				ImageName: "/resources/categories/" + categoryModel.CategoryName + "/products/" + productDetails["name"].(string) + "/images/" + imageName,
			}

			productImages = append(productImages, imageModel)
		}
		productModel.ProductImages = productImages
	}

	return productModel
}

func (productsController *ProductsController) listAllCategories(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(productsController.categoriesArray); err != nil {
		panic(err)
	}
}

func (productsController *ProductsController) listAllProductsByCategory(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	categoryID := r.FormValue("category_id")
	categoryIDi, _ := strconv.Atoi(categoryID)

	productsByCategory := productsController.productsArray[categoryIDi]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(productsByCategory); err != nil {
		panic(err)
	}
}

func (productsController *ProductsController) showProductDetails(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	productID := r.FormValue("product_id")
	productIDi, _ := strconv.Atoi(productID)

	categoryID := productsController.productsInCategory[productIDi]
	productsByCategory := productsController.productsArray[categoryID]

	productModel := &ProductModel{}

	for _, value := range productsByCategory {
		productValueID := (value.(*ProductModel)).ProductID
		if productValueID == productIDi {
			productModel = nil
			productModel = value.(*ProductModel)

			break
		}
	}

	if productModel != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(productModel); err != nil {
			panic(err)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)

	if err := json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"}); err != nil {
		panic(err)
	}
}
