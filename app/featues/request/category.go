package request

type Category struct {
	Name        string `json:"name" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}
