package domain

// Product 定義模型
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
	Stock int     `json:"stock" binding:"required,min=0"`
}

// ProductRepository 定義資料操作的介面
type ProductRepository interface {
	GetAll() []Product
	GetByID(id int) (*Product, error)
	Create(p Product) Product
	Update(id int, p Product) (*Product, error)
	Delete(id int) error
}
