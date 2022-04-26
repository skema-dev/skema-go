package elastic

import (
	"testing"

	"github.com/skema-dev/skema-go/logging"
)

func TestConvertMapToStruct(t *testing.T) {
	type TestData struct {
		Name     string
		Age      int
		Property map[string]string
	}

	value := map[string]interface{}{
		"name": "user1",
		"Age":  10,
		"Property": map[string]string{
			"Property1": "aaaaa",
			"Property2": "bbbbb",
		},
	}

	result := TestData{}

	ConvertMapToStruct(value, &result)

	logging.Infof("%v", result)
}
