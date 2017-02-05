package oauth2

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"regexp"

	"net/url"

	"github.com/Ssawa/LinkLetter/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestGoogleGenerateAuthorizationURL(t *testing.T) {
	google := Google{}

	expected := "https://accounts.google.com/o/oauth2/v2/auth?client_id=1234&prompt=select_account&redirect_uri=https%3A%2F%2Flocalhost.com&response_type=code&scope=email"
	assert.Equal(t, expected, google.GenerateAuthorizationURL("https://localhost.com", "1234", "email"))
}

func TestGoogleExtractAuthorizationCode(t *testing.T) {
	google := Google{}

	accessDeniedReq := httptest.NewRequest("GET", "https://oauth2.example.com/auth?error=access_denied", nil)
	code, err := google.ExtractAuthorizationCode(accessDeniedReq)
	assert.Equal(t, "", code)
	assert.NotNil(t, err)

	noCodeReq := httptest.NewRequest("GET", "https://oauth2.example.com/auth", nil)
	code, err = google.ExtractAuthorizationCode(noCodeReq)
	assert.Equal(t, "", code)
	assert.NotNil(t, err)

	workingReq := httptest.NewRequest("GET", "https://oauth2.example.com/auth?code=4/P7q7W91a-oMsCeLvIaQm6bTrgtp7", nil)
	code, err = google.ExtractAuthorizationCode(workingReq)
	assert.Equal(t, "4/P7q7W91a-oMsCeLvIaQm6bTrgtp7", code)
	assert.Nil(t, err)
}

func TestGoogleGenerateAccessTokenReq(t *testing.T) {
	google := Google{}
	req := google.GenerateAccessTokenRequest("abcd", "https://localhost.com", "1234", "5678")

	dataString, _ := ioutil.ReadAll(req.Body)
	data, err := url.ParseQuery(string(dataString))
	assert.Nil(t, err)
	assert.Equal(t, data.Get("code"), "abcd")
	assert.Equal(t, data.Get("client_id"), "1234")
	assert.Equal(t, data.Get("client_secret"), "5678")
	assert.Equal(t, data.Get("redirect_uri"), "https://localhost.com")
}

func TestGoogleExtractAccessToken(t *testing.T) {
	google := Google{}
	resp := httptest.NewRecorder()
	resp.Code = 200
	resp.WriteString(`
    {
        "access_token":"1/fFAGRNJru1FTz70BzhT3Zg",
        "expires_in":3920,
        "token_type":"Bearer"
    }
    `)

	token, err := google.ExtractAccessToken(resp.Result())
	assert.Nil(t, err)
	assert.Equal(t, "1/fFAGRNJru1FTz70BzhT3Zg", token)
}

func TestGoogleAuthenticate(t *testing.T) {
	transport := testhelpers.FakeTransport(func(req *http.Request) (*http.Response, error) {
		resp := httptest.NewRecorder()
		resp.Code = 200
		resp.WriteString(`
		{
			"hd": "localprojects.com" 
			}
		`)
		return resp.Result(), nil
	})
	defer transport.Close()

	google := Google{}

	pattern, _ := regexp.Compile("localprojects\\.(com|net)")
	auth, err := google.Authenticate("abcd", pattern)
	assert.Nil(t, err)
	assert.True(t, auth)

	pattern, _ = regexp.Compile("helloWorld")
	auth, err = google.Authenticate("abcd", pattern)
	assert.Nil(t, err)
	assert.False(t, auth)
}

func TestGoogleGetProfileData(t *testing.T) {
	transport := testhelpers.FakeTransport(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "Bearer theToken", req.Header.Get("Authorization"))
		resp := httptest.NewRecorder()
		resp.Code = 200
		resp.WriteString(`
		{
			"family_name": "DiMaggio", 
			"name": "Charles DiMaggio", 
			"picture": "https://lh3.googleusercontent.com/-XdUIqdMkCWA/AAAAAAAAAAI/AAAAAAAAAAA/4252rscbv5M/photo.jpg", 
			"locale": "en", 
			"email": "charlesdimaggio@localprojects.com", 
			"given_name": "Charles", 
			"id": "105062841920873375457", 
			"hd": "localprojects.com", 
			"verified_email": true
			}
		`)
		return resp.Result(), nil
	})
	defer transport.Close()

	data, err := getProfileData("theToken")
	assert.Nil(t, err)
	assert.Equal(t, data.Email, "charlesdimaggio@localprojects.com")
	assert.Equal(t, data.HostedDomain, "localprojects.com")
	assert.Equal(t, data.Name, "Charles DiMaggio")
	assert.Equal(t, data.Picture, "https://lh3.googleusercontent.com/-XdUIqdMkCWA/AAAAAAAAAAI/AAAAAAAAAAA/4252rscbv5M/photo.jpg")
	assert.Equal(t, data.VerifiedEmail, true)
}
