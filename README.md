# pfservice
**Products File Service** is a simple micro service which can be used as a backend for a simple products catalog applications. This service is file based where every product is described with simple separete yaml configuration file.

Every category contains products which are folders with product description yaml file and folder with images. Changing product category is very easy with just moving the product folder with everything inside in other category folder and updating the category products yaml configuration file with the new product. After modification we need to restart the service or using the browser to open the url for rebuilding data model using secret key defined in main service configuration yaml file and the service will load the new data model ready for serving the clients.

This service is not yet production ready but it can be freely used as a testing mockup service for mobile / web / desktop apps.
