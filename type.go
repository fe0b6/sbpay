package sbpay

type InitObj struct {
	UserName      string
	Password      string
	OrderNumber   string
	Amount        float64
	ReturnUrl     string
	FailUrl       string
	PageView      string
	IsTesting     bool
	CallbackToken string
}

type AnswerObj struct {
	OrderId      string `json:"orderId"`
	FormUrl      string `json:"formUrl"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}
