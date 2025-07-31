package dto

type CustomerData struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
