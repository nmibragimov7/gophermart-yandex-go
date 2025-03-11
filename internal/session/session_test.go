package session

import (
	"go-musthave-diploma-tpl/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	var userID int64 = 1
	cnf := config.Init()

	ssp := SessionProvider{
		Config: cnf,
	}

	token, err := ssp.CreateToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
