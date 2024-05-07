package utils

import (
	"fmt"

	"github.com/asaskevich/govalidator"
)

// Инициализация валидатора
func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// Выполняет валидацию полей структуры
func ValidateStruct(input any) error {
	_, err := govalidator.ValidateStruct(input)
	if err == nil {
		return nil
	}
	data := ""
	if allErrs, ok := err.(govalidator.Errors); ok {
		for _, fld := range allErrs.Errors() {
			data += fmt.Sprintf("field: %#v\n", fld.Error())
		}
	}
	return fmt.Errorf("error: %s\n%s", err.Error(), data)
}
