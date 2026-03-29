package middleware

import (
	authdtos "backend-sport-team-report-go/internal/modules/auth/application/dtos"

	"github.com/gin-gonic/gin"
)

const authenticatedAccountContextKey = "auth.account"

func SetAuthenticatedAccount(c *gin.Context, account authdtos.AuthenticatedAccount) {
	c.Set(authenticatedAccountContextKey, account)
}

func AuthenticatedAccount(c *gin.Context) (authdtos.AuthenticatedAccount, bool) {
	value, ok := c.Get(authenticatedAccountContextKey)
	if !ok {
		return authdtos.AuthenticatedAccount{}, false
	}

	account, ok := value.(authdtos.AuthenticatedAccount)
	if !ok {
		return authdtos.AuthenticatedAccount{}, false
	}

	return account, true
}
