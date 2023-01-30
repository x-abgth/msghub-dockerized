package models

type AuthErrorModel struct {
	ErrorStr string
}

var errorVal *AuthErrorModel

func InitAuthErrorModel(model AuthErrorModel) *AuthErrorModel {
	errorVal = &AuthErrorModel{
		ErrorStr: model.ErrorStr,
	}
	return errorVal
}

func ReturnAuthErrorModel() *AuthErrorModel {
	return errorVal
}

// phone number validation

type IncorrectPhoneModel struct {
	ErrorStr string
}

var phoneErrorVal *IncorrectPhoneModel

func InitPhoneErrorModel(model IncorrectPhoneModel) *IncorrectPhoneModel {
	phoneErrorVal = &IncorrectPhoneModel{
		ErrorStr: model.ErrorStr,
	}
	return phoneErrorVal
}

func ReturnPhoneErrorModel() *IncorrectPhoneModel {
	return phoneErrorVal
}

// For otp validation page

type IncorrectOtpModel struct {
	ErrorStr    string
	PhoneNumber string
	IsLogin     bool
}

var otpPageData *IncorrectOtpModel

func InitOtpErrorModel(model IncorrectOtpModel) *IncorrectOtpModel {
	otpPageData = &IncorrectOtpModel{
		ErrorStr:    model.ErrorStr,
		PhoneNumber: model.PhoneNumber,
		IsLogin:     model.IsLogin,
	}
	return otpPageData
}

func ReturnOtpErrorModel() *IncorrectOtpModel {
	return otpPageData
}
