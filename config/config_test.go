package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	configData = `
version: skema/v1
kind: app
metadata:
    tpl: grpc-go/grpc-go-4
    project:
        projectName: myproject1
        repoType: github
    service:
        serviceName: service1
    pb: https://github.com/skema-dev/test1/test.proto
    custom:
        aa: bbb
        goPackage: abc/a123/app1
        number1: 1
        number2: 3.1415926535
        bool_1: true
        bool_2: false
        bool_3: True
        bool_4: TRUE
        bool_5: FALSE
        unicode_1: "测试Unicode"
`
)

type configTestSuite struct {
	suite.Suite
}

// SetupSuite ...
func (s *configTestSuite) SetupTest() {
}

// TearDownSuite ...
func (s *configTestSuite) TearDownSuite() {
}

// TestPrivateToken ...
func (s *configTestSuite) TestLoadConfig() {
	conf := NewConfigWithString(configData)

	tests := []struct {
		key    string
		expect string
	}{
		{key: "version", expect: "skema/v1"},
		{key: "kind", expect: "app"},
		{key: "metadata.tpl", expect: "grpc-go/grpc-go-4"},
		{key: "metadata.project.repoType", expect: "github"},
	}

	for _, tt := range tests {
		s.T().Run(tt.key, func(t *testing.T) {
			assert.Equal(s.T(), tt.expect, conf.GetString(tt.key))
		})
	}
}

func (s *configTestSuite) TestLoadSubconfig() {
	conf := NewConfigWithString(configData)
	sub := conf.GetSubConfig("metadata.project")

	tests := []struct {
		key    string
		expect string
	}{
		{key: "projectName", expect: "myproject1"},
		{key: "repoType", expect: "github"},
	}

	for _, tt := range tests {
		s.T().Run(tt.key, func(t *testing.T) {
			assert.Equal(s.T(), tt.expect, sub.GetString(tt.key))
		})
	}
}

func (s *configTestSuite) TestValueTypes() {
	conf := NewConfigWithString(configData)
	sub := conf.GetSubConfig("metadata.custom")

	assert.Equal(s.T(), 1, sub.GetInt("number1"))
	assert.Equal(s.T(), 3.1415926535, sub.GetFloat("number2"))
	assert.Equal(s.T(), true, sub.GetBool("bool_1"))
	assert.Equal(s.T(), false, sub.GetBool("bool_2"))
	assert.Equal(s.T(), true, sub.GetBool("bool_3"))
	assert.Equal(s.T(), true, sub.GetBool("bool_4"))
	assert.Equal(s.T(), false, sub.GetBool("bool_5"))
	assert.Equal(s.T(), "测试Unicode", sub.GetString("unicode_1"))
}

func (s *configTestSuite) TestUnmarshal() {
	type Project struct {
		ProjectName string `json:"projectName"`
		RepoType    string `json:"repoType"`
	}
	type MetaData struct {
		Tpl     string  `json:"tpl"`
		Project Project `json:"project"`
	}

	conf := NewConfigWithString(configData)

	metadata := &MetaData{}
	conf.GetValue("metadata", metadata)

	assert.Equal(s.T(), "grpc-go/grpc-go-4", metadata.Tpl)
	assert.Equal(s.T(), "myproject1", metadata.Project.ProjectName)
	assert.Equal(s.T(), "github", metadata.Project.RepoType)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}
