package requests

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdateCategoryRequest struct {	
	Name     string `json:"name" binding:"required"`
	IsActive *bool   `json:"is_active"`
}

type DeleteCategoryRequest struct {	
	Name     string `json:"name" binding:"required"`
	IsActive *bool   `json:"is_active"`
}