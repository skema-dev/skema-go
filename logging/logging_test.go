package logging_test

import (
	"testing"

	"github.com/skema-dev/skema-go/logging"
	"github.com/stretchr/testify/suite"
)

type loggingTestSuite struct {
	suite.Suite
}

func (s *loggingTestSuite) SetupTest() {
	logging.Init(logging.DebugLevel, "console")
}

func (s *loggingTestSuite) TearDownSuite() {
}

func (s *loggingTestSuite) TestStructuredLogging() {
	logging.Infow("test1",
		"key1", "value1",
		"key2", 2,
	)

	logging.Debugw("test1",
		"key1", "value1",
		"key2", 2,
	)

	logging.Warnw("test1",
		"key1", "value1",
		"key2", 2,
	)

	logging.Errorw("test1",
		"key1", "value1",
		"key2", 2,
	)
}

func (s *loggingTestSuite) TestFormattedLogging() {
	logging.Infof("test1 with %s, %s, %s, %d\n",
		"key1", "value1",
		"key2", 2,
	)
	logging.Debugf("test1 with %s, %s, %s, %d\n",
		"key1", "value1",
		"key2", 2,
	)
	logging.Warnf("test1 with %s, %s, %s, %d\n",
		"key1", "value1",
		"key2", 2,
	)
	logging.Errorf("test1 with %s, %s, %s, %d\n",
		"key1", "value1",
		"key2", 2,
	)

}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(loggingTestSuite))
}
