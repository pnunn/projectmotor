package validator

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-playground/form"
	"net/http"
	"reflect"
)

var decoder *form.Decoder

type Validator interface {
	Validate() error
}

type ValidatedSlice []Validated
type KeyValue struct {
	Key   string
	Value string
}

type Validated struct {
	Key   string
	Value string
	Error string
}

func getKeyValue(val reflect.Value, fieldName string, i int) []KeyValue {
	switch val.Field(i).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []KeyValue{{
			Key:   fieldName,
			Value: fmt.Sprintf("%d", val.Field(i).Int()),
		}}
	case reflect.String:
		return []KeyValue{{
			Key:   fieldName,
			Value: val.Field(i).String(),
		}}
	case reflect.Bool:
		return []KeyValue{{
			Key:   fieldName,
			Value: fmt.Sprintf("%t", val.Field(i).Bool()),
		}}
	}
	return []KeyValue{}
}

func parseErrors(errors validation.Errors, data any) []Validated {
	emap := map[string]string{}
	vmap := []Validated{}
	for k, v := range errors {
		emap[k] = v.Error()
	}
	value := reflect.ValueOf(data).Elem()
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Type().Field(i).Name
		kv := getKeyValue(value, fieldName, i)
		for _, kv := range kv {
			vmap = append(vmap, Validated{
				Key:   kv.Key,
				Value: kv.Value,
				Error: emap[kv.Key],
			})
		}
	}
	return vmap
}

func (vs ValidatedSlice) GetByKey(key string) Validated {
	for _, v := range vs {
		if v.Key == key {
			return v
		}
	}
	return Validated{}
}

func Validate(v Validator, r *http.Request) (bool, ValidatedSlice, error) {
	decoder = form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		return false, []Validated{}, err
	}
	err = decoder.Decode(&v, r.Form)
	if err != nil {
		return false, []Validated{}, err
	}
	err = v.Validate()
	if errors, ok := err.(validation.Errors); ok {
		parsedErrors := parseErrors(errors, v)
		return false, parsedErrors, nil
	}
	if err != nil {
		return false, []Validated{}, err
	}
	return true, []Validated{}, nil
}
