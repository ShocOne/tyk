package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk/header"

	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/gateway"
	"github.com/TykTechnologies/tyk/test"
	"github.com/TykTechnologies/tyk/user"
)

const (
	grpcTestMaxSize = 100000000
	grpcAuthority   = "localhost"

	testHeaderName  = "Testheader"
	testHeaderValue = "testvalue"
)

func loadTestGRPCAPIs(s *gateway.Test) {
	s.Gw.BuildAndLoadAPI(func(spec *gateway.APISpec) {
		spec.APIID = "1"
		spec.OrgID = gateway.MockOrgID
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "authorization",
		}
		spec.UseKeylessAccess = false
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-test-api/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				{Name: "testPreHook1"},
			},
			Driver: apidef.GrpcDriver,
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "2"
		spec.OrgID = gateway.MockOrgID
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "authorization",
		}
		spec.UseKeylessAccess = true
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-test-api-2/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Pre: []apidef.MiddlewareDefinition{
				{Name: "testPreHook2"},
			},
			Driver: apidef.GrpcDriver,
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "3"
		spec.OrgID = "default"
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "authorization",
		}
		spec.UseKeylessAccess = false
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-test-api-3/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Post: []apidef.MiddlewareDefinition{
				{Name: "testPostHook1"},
			},
			Driver: apidef.GrpcDriver,
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "4"
		spec.OrgID = "default"
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "authorization",
		}
		spec.UseKeylessAccess = false
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-test-api-4/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Response: []apidef.MiddlewareDefinition{
				{Name: "testResponseHook"},
			},
			Driver: apidef.GrpcDriver,
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "ignore_plugin"
		spec.OrgID = gateway.MockOrgID
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "authorization",
		}
		spec.UseKeylessAccess = false
		spec.EnableCoProcessAuth = true
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			DefaultVersion: "v1",
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name:             "v1",
					UseExtendedPaths: true,
					ExtendedPaths: apidef.ExtendedPathsSet{
						Ignored: []apidef.EndPointMeta{
							{
								Path:       "/anything",
								IgnoreCase: true,
								MethodActions: map[string]apidef.EndpointMethodMeta{
									http.MethodGet: {
										Action: apidef.NoAction,
										Code:   http.StatusOK,
									},
								},
							},
						},
					},
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-test-api-ignore/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Driver: apidef.GrpcDriver,
			IdExtractor: apidef.MiddlewareIdExtractor{
				ExtractFrom: apidef.HeaderSource,
				ExtractWith: apidef.ValueExtractor,
				ExtractorConfig: map[string]interface{}{
					"header_name": "Authorization",
				},
			},
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "6"
		spec.OrgID = "default"
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "Authorization",
		}
		spec.CustomPluginAuthEnabled = true
		spec.UseKeylessAccess = false
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-auth-hook-test-api-1/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Driver: apidef.GrpcDriver,
			AuthCheck: apidef.MiddlewareDefinition{
				Name: "testAuthHook1",
			},
			IdExtractor: apidef.MiddlewareIdExtractor{
				Disabled:        false,
				ExtractFrom:     apidef.HeaderSource,
				ExtractWith:     apidef.ValueExtractor,
				ExtractorConfig: map[string]interface{}{"header_name": "Authorization"},
			},
		}
	}, func(spec *gateway.APISpec) {
		spec.APIID = "7"
		spec.OrgID = "default"
		spec.Auth = apidef.AuthConfig{
			AuthHeaderName: "Authorization",
		}
		spec.CustomPluginAuthEnabled = true
		spec.UseKeylessAccess = false
		spec.VersionData = struct {
			NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
			DefaultVersion string                        `bson:"default_version" json:"default_version"`
			Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
		}{
			NotVersioned: true,
			Versions: map[string]apidef.VersionInfo{
				"v1": {
					Name: "v1",
				},
			},
		}
		spec.Proxy.ListenPath = "/grpc-auth-hook-test-api-2/"
		spec.Proxy.StripListenPath = true
		spec.CustomMiddleware = apidef.MiddlewareSection{
			Driver: apidef.GrpcDriver,
			AuthCheck: apidef.MiddlewareDefinition{
				Name: "testAuthHook1",
			},
			IdExtractor: apidef.MiddlewareIdExtractor{
				Disabled:        true,
				ExtractFrom:     apidef.HeaderSource,
				ExtractWith:     apidef.ValueExtractor,
				ExtractorConfig: map[string]interface{}{"header_name": "Authorization"},
			},
		}
	},
		func(spec *gateway.APISpec) {
			spec.APIID = "8"
			spec.OrgID = "default"
			spec.Auth = apidef.AuthConfig{
				AuthHeaderName: "Authorization",
			}
			spec.UseKeylessAccess = true
			spec.VersionData = struct {
				NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
				DefaultVersion string                        `bson:"default_version" json:"default_version"`
				Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
			}{
				NotVersioned: true,
				Versions: map[string]apidef.VersionInfo{
					"v1": {
						Name: "v1",
					},
				},
			}
			spec.Proxy.ListenPath = "/grpc-config-data-1/"
			spec.Proxy.StripListenPath = true
			spec.CustomMiddleware = apidef.MiddlewareSection{
				Response: []apidef.MiddlewareDefinition{
					{Name: "testConfigDataResponseHook"},
				},
				Driver: apidef.GrpcDriver,
			}
			spec.ConfigData = map[string]interface{}{"key": "value"}
			spec.ConfigDataDisabled = false
		},
		func(spec *gateway.APISpec) {
			spec.APIID = "9"
			spec.OrgID = "default"
			spec.Auth = apidef.AuthConfig{
				AuthHeaderName: "Authorization",
			}
			spec.UseKeylessAccess = true
			spec.VersionData = struct {
				NotVersioned   bool                          `bson:"not_versioned" json:"not_versioned"`
				DefaultVersion string                        `bson:"default_version" json:"default_version"`
				Versions       map[string]apidef.VersionInfo `bson:"versions" json:"versions"`
			}{
				NotVersioned: true,
				Versions: map[string]apidef.VersionInfo{
					"v1": {
						Name: "v1",
					},
				},
			}
			spec.Proxy.ListenPath = "/grpc-config-data-2/"
			spec.Proxy.StripListenPath = true
			spec.CustomMiddleware = apidef.MiddlewareSection{
				Response: []apidef.MiddlewareDefinition{
					{Name: "testConfigDataResponseHook"},
				},
				Driver: apidef.GrpcDriver,
			}
			spec.ConfigData = map[string]interface{}{"key": "value"}
			spec.ConfigDataDisabled = true
		},
	)
}

func TestMain(m *testing.M) {
	os.Exit(gateway.InitTestMain(context.Background(), m))
}

func TestGRPCDispatch(t *testing.T) {
	ts, cleanupFn := startTestServices(t)
	t.Cleanup(cleanupFn)

	keyID := gateway.CreateSession(ts.Gw, func(s *user.SessionState) {
		s.MetaData = map[string]interface{}{
			"testkey":  map[string]interface{}{"nestedkey": "nestedvalue"},
			"testkey2": "testvalue",
		}
	})
	headers := map[string]string{"authorization": keyID}

	t.Run("Pre Hook with SetHeaders", func(t *testing.T) {
		res, err := ts.Run(t, test.TestCase{
			Path:    "/grpc-test-api/",
			Method:  http.MethodGet,
			Code:    http.StatusOK,
			Headers: headers,
		})
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Couldn't read response body: %s", err.Error())
		}
		var testResponse gateway.TestHttpResponse
		err = json.Unmarshal(data, &testResponse)
		if err != nil {
			t.Fatalf("Couldn't unmarshal test response JSON: %s", err.Error())
		}
		value, ok := testResponse.Headers[testHeaderName]
		if !ok {
			t.Fatalf("Header not found, expecting %s", testHeaderName)
		}
		if value != testHeaderValue {
			t.Fatalf("Header value isn't %s", testHeaderValue)
		}
	})

	t.Run("Pre Hook with UTF-8/non-UTF-8 request data", func(t *testing.T) {
		fileData := gateway.GenerateTestBinaryData()
		var buf bytes.Buffer
		multipartWriter := multipart.NewWriter(&buf)
		file, err := multipartWriter.CreateFormFile("file", "test.bin")
		if err != nil {
			t.Fatalf("Couldn't use multipart writer: %s", err.Error())
		}
		_, err = fileData.WriteTo(file)
		if err != nil {
			t.Fatalf("Couldn't write to multipart file: %s", err.Error())
		}
		field, err := multipartWriter.CreateFormField("testfield")
		if err != nil {
			t.Fatalf("Couldn't use multipart writer: %s", err.Error())
		}
		_, err = field.Write([]byte("testvalue"))
		if err != nil {
			t.Fatalf("Couldn't write to form field: %s", err.Error())
		}
		err = multipartWriter.Close()
		if err != nil {
			t.Fatalf("Couldn't close multipart writer: %s", err.Error())
		}

		ts.Run(t, []test.TestCase{
			{Path: "/grpc-test-api-2/", Code: 200, Data: &buf, Headers: map[string]string{"Content-Type": multipartWriter.FormDataContentType()}},
			{Path: "/grpc-test-api-2/", Code: 200, Data: "{}", Headers: map[string]string{"Content-Type": "application/json"}},
		}...)
	})

	t.Run("Post Hook with metadata", func(t *testing.T) {
		ts.Run(t, test.TestCase{
			Path:    "/grpc-test-api-3/",
			Method:  http.MethodGet,
			Code:    http.StatusOK,
			Headers: headers,
		})
	})

	t.Run("Response hook", func(t *testing.T) {
		ts.Run(t, test.TestCase{
			Path:      "/grpc-test-api-4/",
			Method:    http.MethodGet,
			Code:      http.StatusOK,
			Headers:   headers,
			BodyMatch: "newbody",
			HeadersMatch: map[string]string{
				header.ContentLength: strconv.Itoa(len("newbody")),
			},
		})
	})

	t.Run("Post Hook with allowed message length", func(t *testing.T) {
		test.Flaky(t)

		s := randStringBytes(20000000)
		ts.Run(t, test.TestCase{
			Path:    "/grpc-test-api-3/",
			Method:  http.MethodGet,
			Code:    http.StatusOK,
			Headers: headers,
			Data:    s,
		})
	})

	t.Run("Post Hook with with unallowed message length", func(t *testing.T) {
		test.Flaky(t)

		s := randStringBytes(150000000)
		ts.Run(t, test.TestCase{
			Path:    "/grpc-test-api-3/",
			Method:  http.MethodGet,
			Code:    http.StatusInternalServerError,
			Headers: headers,
			Data:    s,
		})
	})
}

func BenchmarkGRPCDispatch(b *testing.B) {
	ts, cleanupFn := startTestServices(b)
	b.Cleanup(cleanupFn)

	keyID := gateway.CreateSession(ts.Gw)
	headers := map[string]string{"authorization": keyID}

	b.Run("Pre Hook with SetHeaders", func(b *testing.B) {
		basepath := "/grpc-test-api/"
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ts.Run(b, test.TestCase{
				Path:    basepath,
				Method:  http.MethodGet,
				Code:    http.StatusOK,
				Headers: headers,
			})
		}
	})
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	return strings.Repeat(string(letters[rand.Intn(len(letters))]), n)
}

func TestGRPCIgnore(t *testing.T) {
	ts, cleanupFn := startTestServices(t)
	t.Cleanup(cleanupFn)

	basepath := "/grpc-test-api-ignore/"

	// no header
	ts.Run(t, test.TestCase{
		Path:   basepath + "something",
		Method: http.MethodGet,
		Code:   http.StatusBadRequest,
		BodyMatchFunc: func(b []byte) bool {
			return bytes.Contains(b, []byte("Authorization field missing"))
		},
	})

	ts.Run(t, test.TestCase{
		Path:   basepath + "anything",
		Method: http.MethodGet,
		Code:   http.StatusOK,
	})

	// bad header
	headers := map[string]string{"authorization": "bad"}
	ts.Run(t, test.TestCase{
		Path:    basepath + "something",
		Method:  http.MethodGet,
		Code:    http.StatusForbidden,
		Headers: headers,
	})

	ts.Run(t, test.TestCase{
		Path:    basepath + "anything",
		Method:  http.MethodGet,
		Code:    http.StatusOK,
		Headers: headers,
	})
}

func TestGRPCAuthHook(t *testing.T) {
	ts, cleanupFn := startTestServices(t)
	t.Cleanup(cleanupFn)

	t.Run("id extractor enabled", func(t *testing.T) {
		basepath := "/grpc-auth-hook-test-api-1/"
		spec := &gateway.APISpec{
			APIDefinition: &apidef.APIDefinition{
				OrgID: "default",
			},
		}
		baseMW := gateway.NewBaseMiddleware(ts.Gw, spec, nil, nil)
		baseExtractor := gateway.BaseExtractor{
			BaseMiddleware: baseMW,
		}
		expectedSessionID := baseExtractor.GenerateSessionID("abc", baseMW)
		_, _ = ts.Run(t, []test.TestCase{
			{Method: http.MethodGet, Path: basepath, Headers: map[string]string{"Authorization": "abc"}, Code: http.StatusOK},
			{Method: http.MethodGet, Path: fmt.Sprintf("/tyk/keys/%s", expectedSessionID), AdminAuth: true, Code: http.StatusOK},
		}...)
	})

	// won't extract id and a session with sessionID as token is created
	t.Run("id extractor disabled", func(t *testing.T) {
		basepath := "/grpc-auth-hook-test-api-2/"
		_, _ = ts.Run(t, []test.TestCase{
			{Method: http.MethodGet, Path: basepath, Headers: map[string]string{"Authorization": "abc"}, Code: http.StatusOK},
			{Method: http.MethodGet, Path: "/tyk/keys/abc", AdminAuth: true, Code: http.StatusOK},
		}...)
	})
}

func TestGRPC_MultiAuthentication(t *testing.T) {
	ts, cleanupFn := startTestServices(t)
	t.Cleanup(cleanupFn)

	const (
		apiID          = "my-api-id"
		sessionMetaKey = "sessionMetaKey"

		customAuthSessionMetaValue = "customAuthSessionMetaValue"
		customAuthSessionID        = "abc"
		customAuthSessionRate      = 100

		authTokenSessionMetaValue = "authTokenSessionMetaValue"
		authTokenSessionRate      = 200
	)

	api := gateway.BuildAPI(func(spec *gateway.APISpec) {
		spec.APIID = apiID
		spec.Proxy.ListenPath = "/"
		spec.UseKeylessAccess = false
		spec.EnableCoProcessAuth = true
		spec.AuthConfigs = map[string]apidef.AuthConfig{
			apidef.AuthTokenType: {
				AuthHeaderName: "AuthToken",
			},
		}
		spec.UseStandardAuth = true
		spec.UseKeylessAccess = false
		spec.VersionData.Versions["v1"] = apidef.VersionInfo{
			GlobalResponseHeaders: map[string]string{
				sessionMetaKey: "$tyk_meta." + sessionMetaKey,
			},
		}
		spec.CustomMiddleware.Driver = apidef.GrpcDriver
		spec.CustomMiddleware.AuthCheck.Name = "testAuthHook1"
		spec.CustomMiddleware.IdExtractor.Extractor = nil
	})[0]

	_, authTokenSessionID := ts.CreateSession(func(s *user.SessionState) {
		s.Rate = authTokenSessionRate
		s.MetaData = map[string]interface{}{
			sessionMetaKey: authTokenSessionMetaValue,
		}
		s.AccessRights = map[string]user.AccessDefinition{apiID: {
			APIID: apiID, Versions: []string{"v1"},
		}}
	})

	check := func(t *testing.T, baseIdentityProvidedBy apidef.AuthTypeEnum, keyName string, headerVal string, rate int) {
		t.Helper()

		api.BaseIdentityProvidedBy = baseIdentityProvidedBy
		ts.Gw.LoadAPI(api)

		_, _ = ts.Run(t, []test.TestCase{
			{Headers: map[string]string{"Authorization": customAuthSessionID, "AuthToken": authTokenSessionID},
				HeadersMatch: map[string]string{sessionMetaKey: headerVal}, Code: http.StatusOK},
		}...)

		retSession, found := ts.Gw.GlobalSessionManager.SessionDetail(api.OrgID, keyName, false)
		assert.Equal(t, float64(rate), retSession.Rate)
		assert.True(t, found)
	}

	t.Run("custom base identity", func(t *testing.T) {
		check(t, apidef.CustomAuth, customAuthSessionID, customAuthSessionMetaValue, customAuthSessionRate)
	})

	t.Run("auth token base identity", func(t *testing.T) {
		check(t, apidef.AuthToken, authTokenSessionID, authTokenSessionMetaValue, authTokenSessionRate)
	})
}

func TestGRPCConfigData(t *testing.T) {
	ts, cleanupFn := startTestServices(t)
	t.Cleanup(cleanupFn)

	t.Run("config data disabled", func(t *testing.T) {
		basepath := "/grpc-config-data-1/"
		_, _ = ts.Run(t, []test.TestCase{
			{Method: http.MethodGet, Path: basepath, Code: http.StatusOK,
				HeadersMatch: map[string]string{"x-config-data": "true"},
			},
		}...)
	})

	t.Run("config data disabled", func(t *testing.T) {
		basepath := "/grpc-config-data-2/"
		_, _ = ts.Run(t, []test.TestCase{
			{Method: http.MethodGet, Path: basepath, Code: http.StatusOK,
				HeadersMatch: map[string]string{"x-config-data": "false"},
			},
		}...)
	})

}
