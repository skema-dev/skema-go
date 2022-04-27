package elastic

import (
	"testing"

	"github.com/skema-dev/skema-go/logging"
	"github.com/stretchr/testify/assert"
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

func TestCreateOrderCondition(t *testing.T) {
	s := "id desc"
	sort := createSortCondition(s)
	assert.Equal(t, 1, len(sort))
	assert.Equal(t, "desc", sort[0]["id"])

	s = "id desc   "
	sort = createSortCondition(s)
	assert.Equal(t, 1, len(sort))
	assert.Equal(t, "desc", sort[0]["id"])

	s = "id desc  ,    name asc"
	sort = createSortCondition(s)
	assert.Equal(t, 2, len(sort))
	assert.Equal(t, "desc", sort[0]["id"])
	assert.Equal(t, "asc", sort[1]["name"])

	s = "id   ,    name   "
	sort = createSortCondition(s)
	assert.Equal(t, 2, len(sort))
	assert.Equal(t, "desc", sort[0]["id"])
	assert.Equal(t, "desc", sort[1]["name"])
}
