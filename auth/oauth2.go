package auth

import (
	"context"
	"time"

	"github.com/HuaTug/My-RPC/codes"
	"github.com/HuaTug/My-RPC/interceptor"
	"golang.org/x/oauth2"
)

type oAuth2 struct {
	token *oauth2.Token
}

func (o *oAuth2) AuthType() string {
	return "oauth2"
}

// NewOAuth2ByToken supports the generation of an oauth2 based on a string token
func NewOAuth2ByToken(token string) *oAuth2 {
	return &oAuth2{
		token: &oauth2.Token{
			AccessToken: token,
			Expiry:      time.Now().Add(time.Hour),
		},
	}
}

// NewOAuth2 supports the generation of an oauth2 based on an oauth2 token
func NewOAuth2(t *oauth2.Token) *oAuth2 {
	return &oAuth2{
		token: t,
	}
}

// return the token format "Bear <access-token>"
func (o *oAuth2) GetMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

	if o.token == nil {
		return nil, codes.ClientCertFailError
	}

	return map[string]string{
		"authorization": o.token.Type() + " " + o.token.AccessToken,
	}, nil
}

// AuthFunc verifies that the token is valid or not
// This achieves a pluggable extension of function types
type AuthFunc func(ctx context.Context) (context.Context, error)

// BuildAuthFilter constructs a client interceptor with an AuthFunc
func BuildAuthInterceptor(af AuthFunc) interceptor.ServerInterceptor {

	return func(ctx context.Context, req interface{}, handler interceptor.Handler) (interface{}, error) {
		newCtx, err := af(ctx)

		if err != nil {
			return nil, codes.NewFrameworkError(codes.ClientCertFail, err.Error())
		}

		return handler(newCtx, req)
	}
}
