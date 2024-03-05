package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v3"
	"github.com/jjonline/go-lib-backend/guzzle"
	"github.com/jjonline/go-lib-backend/logger"
	"github.com/jjonline/serve-swagger-ui/client"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/jjonline/serve-swagger-ui/define"
	"github.com/jjonline/serve-swagger-ui/utils/memory"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var signer jose.Signer

var (
	AuthorizationHasExpiredOrInvalid = errors.New("authorization has expired or invalid")
	AuthorizationParamInvalid        = errors.New("callback URL needed params is missing")
	LoginFailed                      = errors.New("some error occurred, login failed")
	EmailIsNotAllow                  = errors.New("your email account is not allowed to log in")
)

// Token request token
type Token struct {
	Email         string `json:"email"`
	Authenticated bool   `json:"authenticated"`
	ShouldLogin   bool   `json:"should_login"`
}

// JwtToken jwt token struct
type JwtToken struct {
	Sub string `json:"sub"` // JWT subject, just use oauth email, such as google email address
	Exp int64  `json:"exp"` // JWT expired unix timestamp
	Iat int64  `json:"iat"` // JWT issue unix timestamp
}

// googleAccessToken google oauth accessToken structure
type googleAccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// googleUserInfo google oauth user info
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

// microsoftAccessToken microsoft oauth accessToken structure
type microsoftAccessToken struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// microsoftUserInfo microsoft oauth user info
type microsoftUserInfo struct {
	//OdataContext      string      `json:"@odata.context"`
	ID string `json:"id"`
	//BusinessPhones    []string    `json:"businessPhones"`
	DisplayName       string      `json:"displayName"`
	GivenName         string      `json:"givenName"`
	JobTitle          string      `json:"jobTitle"`
	Mail              interface{} `json:"mail"` // null or a string
	MobilePhone       string      `json:"mobilePhone"`
	OfficeLocation    string      `json:"officeLocation"`
	PreferredLanguage interface{} `json:"preferredLanguage"`
	Surname           string      `json:"surname"`
	UserPrincipalName string      `json:"userPrincipalName"` // email
}

// oauthService oAuth service
type oauthService struct{}

// region google oauth

// GoogleRedirectURL get google oauth login redirect URL
func (o *oauthService) GoogleRedirectURL(ctx *gin.Context) string {
	param := make(url.Values, 0)
	param.Set("scope", "https://www.googleapis.com/auth/userinfo.email")
	param.Set("response_type", "code")
	param.Set("state", o.makeState(ctx))
	param.Set("redirect_uri", o.makeGoogleCallback())
	param.Set("client_id", conf.Config.Google.ClientID)
	return guzzle.ToQueryURL("https://accounts.google.com/o/oauth2/v2/auth", param)
}

// GoogleCallback handle google oauth login callback
func (o *oauthService) GoogleCallback(ctx *gin.Context) error {
	// ① check state remission CSRF
	state := ctx.Query("state")
	code := ctx.Query("code")
	if code == "" || state == "" || state != o.getState(state) {
		return AuthorizationParamInvalid
	}

	// ② code exchange accessToken
	accessToken, err := o.googleCode2accessToken(code)
	if err != nil {
		return err
	}

	// ③ accessToken exchange user info
	email, err := o.googleAccessToken2Email(accessToken)
	if err != nil {
		return err
	}

	// ④ check can attempt login
	if !o.canAttempt(email) {
		return EmailIsNotAllow
	}

	// ⑤ generate JWT then set cookie
	jwt, err := o.generateJwt(email)
	if err != nil {
		return LoginFailed
	}
	o.SetCookie(ctx, jwt)

	return nil
}

// googleCode2accessToken use oauth code exchange accessToken
func (o *oauthService) googleCode2accessToken(code string) (string, error) {
	param := make(url.Values, 0)
	param.Set("client_id", conf.Config.Google.ClientID)
	param.Set("client_secret", conf.Config.Google.ClientSecret)
	param.Set("code", code)
	param.Set("grant_type", "authorization_code")
	param.Set("redirect_uri", o.makeGoogleCallback())

	result, err := client.Guzzle.PostForm(context.TODO(), "https://oauth2.googleapis.com/token", param, nil)
	if err != nil {
		return "", err
	}

	var accessToken = googleAccessToken{}
	if err := json.Unmarshal(result.Body, &accessToken); err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil
}

// googleAccessToken2UserInfo use google oAuth accessToken get user info
func (o *oauthService) googleAccessToken2Email(accessToken string) (string, error) {
	param := make(url.Values, 0)
	param.Set("access_token", accessToken)
	result, err := client.Guzzle.Get(context.TODO(), "https://www.googleapis.com/oauth2/v2/userinfo", param, nil)
	if err != nil {
		return "", err
	}

	var user = googleUserInfo{}
	if err := json.Unmarshal(result.Body, &user); err != nil {
		return "", err
	}
	return user.Email, nil
}

// endregion

// region microsoft oauth

// MicrosoftRedirectURL get Microsoft oauth login redirect URL
func (o *oauthService) MicrosoftRedirectURL(ctx *gin.Context) string {
	param := make(url.Values, 0)
	param.Set("client_id", conf.Config.Microsoft.ClientID)
	param.Set("response_type", "code")
	param.Set("redirect_uri", o.makeMicrosoftCallback())
	param.Set("scope", "User.Read Mail.Read")
	param.Set("response_mode", "query")
	param.Set("state", o.makeState(ctx))

	baseURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", conf.Config.Microsoft.Tenant)
	return guzzle.ToQueryURL(baseURL, param)
}

// MicrosoftCallback handle Microsoft oauth login callback
func (o *oauthService) MicrosoftCallback(ctx *gin.Context) error {
	// ① check state remission CSRF
	state := ctx.Query("state")
	code := ctx.Query("code")
	if code == "" || state == "" || state != o.getState(state) {
		return AuthorizationParamInvalid
	}

	// ② code exchange accessToken
	accessToken, err := o.microsoftCode2accessToken(code)
	if err != nil {
		return err
	}

	// ③ accessToken exchange user info
	email, err := o.microsoftAccessToken2Email(accessToken)
	if err != nil {
		return err
	}

	// ④ check can attempt login
	if !o.canAttempt(email) {
		return EmailIsNotAllow
	}

	// ⑤ generate JWT then set cookie
	jwt, err := o.generateJwt(email)
	if err != nil {
		return LoginFailed
	}
	o.SetCookie(ctx, jwt)

	return nil
}

// microsoftCode2accessToken use oauth code exchange accessToken
func (o *oauthService) microsoftCode2accessToken(code string) (string, error) {
	param := make(url.Values, 0)
	param.Set("client_id", conf.Config.Microsoft.ClientID)
	param.Set("client_secret", conf.Config.Microsoft.ClientSecret)
	param.Set("code", code)
	param.Set("grant_type", "authorization_code")
	param.Set("redirect_uri", o.makeMicrosoftCallback())

	baseURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", conf.Config.Microsoft.Tenant)
	result, err := client.Guzzle.PostForm(context.TODO(), baseURL, param, nil)
	if err != nil {
		return "", err
	}

	var accessToken = microsoftAccessToken{}
	if err := json.Unmarshal(result.Body, &accessToken); err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil
}

// microsoftAccessToken2Email use microsoft oAuth accessToken get user info
func (o *oauthService) microsoftAccessToken2Email(accessToken string) (string, error) {
	head := make(map[string]string, 0)
	head["Authorization"] = "Bearer " + accessToken

	result, err := client.Guzzle.Get(context.TODO(), "https://graph.microsoft.com/v1.0/me", nil, head)
	if err != nil {
		return "", err
	}

	var user = microsoftUserInfo{}
	if err := json.Unmarshal(result.Body, &user); err != nil {
		return "", err
	}

	// select email column
	if user.Mail != nil {
		return user.Mail.(string), nil
	}
	return user.UserPrincipalName, nil
}

// endregion

// CheckAuthorization check cookie state
func (o *oauthService) CheckAuthorization(ctx *gin.Context) (token Token) {
	token.ShouldLogin = conf.Config.ShouldLogin // set should log in
	if !conf.Config.ShouldLogin {
		return token
	}

	cookie, err := ctx.Cookie(define.AuthCookieName)
	if err != nil {
		return token
	}

	jwtToken, err := o.verifyJwt(cookie)
	if err != nil {
		return token
	}

	// check account can attempt
	if !o.canAttempt(jwtToken.Sub) {
		return token
	}

	token.Email = jwtToken.Sub
	token.Authenticated = true
	return token
}

// CheckIsLoginUsingToken check login status helper
func (o *oauthService) CheckIsLoginUsingToken(ctx *gin.Context) bool {
	if tokenInter, exist := ctx.Get("token"); exist {
		if token, ok := tokenInter.(Token); ok && token.Authenticated {
			return true
		}
	}
	return false
}

// SetCookie set cookie
func (o *oauthService) SetCookie(ctx *gin.Context, jwt string) {
	host, _ := url.Parse(conf.Config.Server.BaseURL)
	expire := int(conf.Config.Server.JwtExpiredTime)
	ctx.SetSameSite(http.SameSiteLaxMode) // set cookie sameSite lax model
	ctx.SetCookie(define.AuthCookieName, jwt, expire, "/", host.Host, host.Scheme == "https", true)
}

// DeleteCookie delete cookie
func (o *oauthService) DeleteCookie(ctx *gin.Context) {
	host, _ := url.Parse(conf.Config.Server.BaseURL)
	ctx.SetSameSite(http.SameSiteLaxMode) // set cookie sameSite lax model
	ctx.SetCookie(define.AuthCookieName, "deleted", -1, "/", host.Host, host.Scheme == "https", true)
}

// region base oauth method

// makeGoogleCallback make google callback URL, which query name is redirect_uri
func (o *oauthService) makeGoogleCallback() string {
	return conf.Config.Server.BaseURL + define.GoogleCallbackRoute
}

// makeMicrosoftCallback make google callback URL, which query name is redirect_uri
func (o *oauthService) makeMicrosoftCallback() string {
	return conf.Config.Server.BaseURL + define.MicrosoftCallbackRoute
}

// makeState make oauth state random string
func (o *oauthService) makeState(ctx *gin.Context) string {
	requestId := logger.GetRequestID(ctx)
	_ = memory.Set(define.GoogleOauthStateCachePrefixKey+requestId, requestId, 5*time.Minute)
	return requestId
}

// getState get local memory cached state,if not exist will return empty string
func (o *oauthService) getState(state string) string {
	reqID := memory.Pull(define.GoogleOauthStateCachePrefixKey + state)
	if reqID == nil {
		return ""
	}
	return reqID.(string)
}

// endregion

// canAttempt check email can attempt auth
func (o *oauthService) canAttempt(email string) bool {
	// ① check specify email
	for _, item := range conf.Config.Account.Email {
		if item == email {
			return true
		}
	}

	// ② check email suffix domain
	emailItems := strings.Split(email, "@")
	if len(emailItems) == 2 {
		for _, item := range conf.Config.Account.Domain {
			if item == emailItems[1] {
				return true
			}
		}
	}
	return false
}

// generateJwt use auth email generate jwt
// email ex: google email address
func (o *oauthService) generateJwt(email string) (jwt string, err error) {
	// init signer
	if signer == nil {
		signer, err = jose.NewSigner(jose.SigningKey{
			Algorithm: "HS256",
			Key:       []byte(conf.Config.Server.JwtKey),
		}, nil)
		if err != nil {
			return
		}
	}

	// construct token struct
	now := time.Now()
	token := JwtToken{
		Sub: email,
		Exp: now.Add(time.Duration(conf.Config.Server.JwtExpiredTime) * time.Second).Unix(),
		Iat: now.Unix(),
	}

	obj, err := json.Marshal(token)
	if err != nil {
		return
	}

	jwsObj, err := signer.Sign(obj)
	if err != nil {
		return
	}

	return jwsObj.CompactSerialize()
}

// verifyJwt verify JWT string
func (o *oauthService) verifyJwt(jwt string) (token JwtToken, err error) {
	var sig *jose.JSONWebSignature
	if sig, err = jose.ParseSigned(jwt); err != nil {
		return JwtToken{}, err
	}

	if err = json.Unmarshal(sig.UnsafePayloadWithoutVerification(), &token); err != nil {
		return JwtToken{}, err
	}

	// validate expired time
	if token.Exp <= time.Now().Unix() {
		return JwtToken{}, AuthorizationHasExpiredOrInvalid
	}

	// check key
	if _, err = sig.Verify([]byte(conf.Config.Server.JwtKey)); err != nil {
		return JwtToken{}, AuthorizationHasExpiredOrInvalid
	}

	return token, nil
}
