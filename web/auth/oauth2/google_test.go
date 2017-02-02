package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"regexp"

	"github.com/Ssawa/LinkLetter/web"
	"github.com/stretchr/testify/assert"
)

func TestGoogleGenerateAuthorizationURL(t *testing.T) {
	google := Google{
		clientID:     "1234",
		clientSecret: "5678",
	}

	expected := "https://accounts.google.com/o/oauth2/v2/auth?client_id=1234&prompt=select_account&redirect_uri=https%3A%2F%2Flocalhost.com&response_type=code&scope=email"
	assert.Equal(t, expected, google.GenerateAuthorizationURL("https://localhost.com").String())
}

func TestGoogleExtractAuthorizationCode(t *testing.T) {
	google := Google{
		clientID:     "1234",
		clientSecret: "5678",
	}

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

func TestGoogleGenerateAccessTokenURL(t *testing.T) {
	google := Google{
		clientID:     "1234",
		clientSecret: "5678",
	}
	expected := "https://www.googleapis.com/oauth2/v4/token?client_id=1234&client_secret=1234&code=authcode&grant_type=authorization_code&redirect_uri=https%3A%2F%2Flocalhost.com"
	assert.Equal(t, expected, google.GenerateAccessTokenURL("authcode", "https://localhost.com").String())
}

func TestGoogleExtractAccessToken(t *testing.T) {
	google := Google{
		clientID:     "1234",
		clientSecret: "5678",
	}
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
	transport := web.FakeTransport(func(req *http.Request) (*http.Response, error) {
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

	google := Google{
		clientID:     "1234",
		clientSecret: "5678",
	}

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
	transport := web.FakeTransport(func(req *http.Request) (*http.Response, error) {
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
