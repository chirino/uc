package utils_test

import (
	"fmt"
	"github.com/chirino/uc/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiError(t *testing.T) {
	assert := assert.New(t)
	e1 := fmt.Errorf("can't read file")
	e2 := fmt.Errorf("can't write file")
	e3 := utils.Errors(e1, e2)
	assert.Equal(`2 errors: #1: can't read file, #2: can't write file`, e3.Error())

	e4 := utils.Errors(e1)
	assert.Equal(e1.Error(), e4.Error())

	e5 := utils.Errors()
	assert.Nil(e5)
}
