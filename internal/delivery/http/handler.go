package http

import (
	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Svc service.ProductService
}

// GetAll doc
// @Summary      Get all products
// @Description  Retrieve all products from the database
// @Tags         products
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Product
// @Router       /products [get]
func (h *ProductHandler) GetAll(c *gin.Context) {
	c.JSON(http.StatusOK, h.Svc.GetAllProducts())
}

// GetByID godoc
// @Summary      Get product by ID
// @Description  Retrieve a specific product by its ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  domain.Product
// @Failure      404  {object}  domain.AppError
// @Router       /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}
	p, err := h.Svc.GetProduct(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// Create godoc
// @Summary      Create a new product
// @Description  Create a new product in the database
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        product  body      domain.Product  true  "Product data"
// @Success      201      {object}  domain.Product
// @Failure      400      {object}  domain.AppError
// @Router       /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}
	newP, err := h.Svc.CreateProduct(p)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, newP)
}

// Update godoc
// @Summary      Update a product
// @Description  Update an existing product by its ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int             true  "Product ID"
// @Param        product  body      domain.Product  true  "Updated product data"
// @Success      200      {object}  domain.Product
// @Failure      400      {object}  domain.AppError
// @Failure      404      {object}  domain.AppError
// @Router       /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}
	updatedP, err := h.Svc.UpdateProduct(id, p)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, updatedP)
}

// Delete godoc
// @Summary      Delete a product
// @Description  Delete a product by its ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Product ID"
// @Success      204
// @Failure      404  {object}  domain.AppError
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrInvalidInput)
		return
	}
	if err := h.Svc.DeleteProduct(id); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
