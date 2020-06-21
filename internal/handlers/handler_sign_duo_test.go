package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/mocks"
)

type SecondFactorDuoPostSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorDuoPostSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	s.mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (s *SecondFactorDuoPostSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndAllowAccess() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Eq(values), s.mock.Ctx).Return(&response, nil)

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	assert.Equal(s.T(), s.mock.Ctx.Response.StatusCode(), 200)
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndDenyAccess() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	response := duo.Response{}
	response.Response.Result = "deny"

	duoMock.EXPECT().Call(gomock.Eq(values), s.mock.Ctx).Return(&response, nil)

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	assert.Equal(s.T(), s.mock.Ctx.Response.StatusCode(), 401)
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndFail() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	duoMock.EXPECT().Call(gomock.Eq(values), s.mock.Ctx).Return(nil, fmt.Errorf("Connnection error"))

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")
}

func (s *SecondFactorDuoPostSuite) TestShouldRedirectUserToDefaultURL() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Any(), s.mock.Ctx).Return(&response, nil)

	s.mock.Ctx.Configuration.DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(signDuoRequestBody{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURL,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldNotReturnRedirectURL() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Any(), s.mock.Ctx).Return(&response, nil)

	bodyBytes, err := json.Marshal(signDuoRequestBody{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *SecondFactorDuoPostSuite) TestShouldRedirectUserToSafeTargetURL() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Any(), s.mock.Ctx).Return(&response, nil)

	bodyBytes, err := json.Marshal(signDuoRequestBody{
		TargetURL: "https://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "https://mydomain.local",
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldNotRedirectToUnsafeURL() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Any(), s.mock.Ctx).Return(&response, nil)

	bodyBytes, err := json.Marshal(signDuoRequestBody{
		TargetURL: "http://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *SecondFactorDuoPostSuite) TestShouldRegenerateSessionForPreventingSessionFixation() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	response := duo.Response{}
	response.Response.Result = testResultAllow

	duoMock.EXPECT().Call(gomock.Any(), s.mock.Ctx).Return(&response, nil)

	bodyBytes, err := json.Marshal(signDuoRequestBody{
		TargetURL: "http://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)

	s.Assert().NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))
}

func TestRunSecondFactorDuoPostSuite(t *testing.T) {
	s := new(SecondFactorDuoPostSuite)
	suite.Run(t, s)
}
