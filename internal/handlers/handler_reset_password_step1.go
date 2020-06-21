package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/session"
)

func identityRetrieverFromStorage(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
	var requestBody resetPasswordStep1RequestBody
	err := json.Unmarshal(ctx.PostBody(), &requestBody)

	if err != nil {
		return nil, err
	}

	details, err := ctx.Providers.UserProvider.GetDetails(requestBody.Username)

	if err != nil {
		return nil, err
	}

	if len(details.Emails) == 0 {
		return nil, fmt.Errorf("User %s has no email address configured", requestBody.Username)
	}

	return &session.Identity{
		Username: requestBody.Username,
		Email:    details.Emails[0],
	}, nil
}

// ResetPasswordIdentityStart the handler for initiating the identity validation for resetting a password.
// We need to ensure the attacker cannot perform user enumeration by always replying with 200 whatever what happens in backend.
var ResetPasswordIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:             "Reset your password",
	MailButtonContent:     "Reset",
	TargetEndpoint:        "/reset-password/step2",
	ActionClaim:           ResetPasswordAction,
	IdentityRetrieverFunc: identityRetrieverFromStorage,
})

func resetPasswordIdentityFinish(ctx *middlewares.AutheliaCtx, username string) {
	userSession := ctx.GetSession()
	// TODO(c.michaud): use JWT tokens to expire the request in only few seconds for better security.
	userSession.PasswordResetUsername = &username
	ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	ctx.ReplyOK()
}

// ResetPasswordIdentityFinish the handler for finishing the identity validation.
var ResetPasswordIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{ActionClaim: ResetPasswordAction}, resetPasswordIdentityFinish)
