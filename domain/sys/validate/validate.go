package validate

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslator "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
	"reflect"
	"strings"
)

var validate *validator.Validate

var translator ut.Translator

func init() {
	validate = validator.New()

	translator, _ := ut.New(en.New(), en.New()).GetTranslator("en")
	enTranslator.RegisterDefaultTranslations(validate, translator)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Check(val any) error {
	if err := validate.Struct(val); err != nil {

		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		var fields FieldErrors
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(translator),
			}
			fields = append(fields, field)
		}
		return fields
	}
	return nil
}

func GenerateUID() string {
	return uuid.NewString()
}

func CheckID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}
	return nil
}
