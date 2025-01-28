package config

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestConfigWithOverrides testing whether ImmichUrl and ImmichApiKey are immutable
func TestImmichUrlImmichApiKeyImmutability(t *testing.T) {

	originalUrl := "https://my-server.com"
	originalApi := "123456"

	c := New()
	c.ImmichUrl = originalUrl
	c.ImmichApiKey = originalApi

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	q := req.URL.Query()
	q.Add("immich_url", "https://my-new-server.com")
	q.Add("immich_api_key", "9999")

	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	err := c.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	assert.Equal(t, originalUrl, c.ImmichUrl, "ImmichUrl field was allowed to be changed")
	assert.Equal(t, originalApi, c.ImmichApiKey, "ImmichApiKey field was allowed to be changed")
}

// TestImmichUrlImmichMulitplePerson tests the addition of multiple persons to the config
func TestImmichUrlImmichMulitplePerson(t *testing.T) {
	c := New()

	e := echo.New()

	q := make(url.Values)
	q.Add("person", "bea")
	q.Add("person", "laura")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err := c.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	assert.Equal(t, 2, len(c.Person), "Expected 2 people to be added")
	assert.Contains(t, c.Person, "bea", "Expected 'bea' to be added to Person slice")
	assert.Contains(t, c.Person, "laura", "Expected 'laura' to be added to Person slice")
}

// TestMalformedURLs testing urls without scheme or ports
func TestMalformedURLs(t *testing.T) {

	var tests = []struct {
		KIOSK_IMMICH_URL string
		Want             string
	}{
		{KIOSK_IMMICH_URL: "nope", Want: defaultScheme + "nope"},
		{KIOSK_IMMICH_URL: "192.168.1.1", Want: defaultScheme + "192.168.1.1"},
		{KIOSK_IMMICH_URL: "192.168.1.1:1234", Want: defaultScheme + "192.168.1.1:1234"},
		{KIOSK_IMMICH_URL: "https://192.168.1.1:1234", Want: "https://192.168.1.1:1234"},
		{KIOSK_IMMICH_URL: "nope:32", Want: defaultScheme + "nope:32"},
	}

	for _, test := range tests {

		t.Run(test.KIOSK_IMMICH_URL, func(t *testing.T) {
			t.Setenv("KIOSK_IMMICH_URL", test.KIOSK_IMMICH_URL)
			t.Setenv("KIOSK_IMMICH_API_KEY", "12345")

			c := New()

			err := c.Load()
			assert.NoError(t, err, "Config load should not return an error")

			assert.Equal(t, test.Want, c.ImmichUrl, "ImmichUrl should be formatted correctly")
		})
	}
}

// TestImmichUrlImmichMulitpleAlbum tests the addition and overriding of multiple albums in the config
func TestImmichUrlImmichMulitpleAlbum(t *testing.T) {

	// configWithBase
	configWithBase := New()
	configWithBase.Album = []string{"BASE_ALBUM"}

	e := echo.New()

	q := make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err := configWithBase.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithBase.Album)

	assert.NotContains(t, configWithBase.Album, "BASE_ALBUM", "BASE_ALBUM should not be present")
	assert.Equal(t, 2, len(configWithBase.Album), "Expected 2 albums to be added")
	assert.Contains(t, configWithBase.Album, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithBase.Album, "ALBUM_2", "ALBUM_2 should be present")

	// configWithoutBase
	configWithoutBase := New()

	q = make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err = configWithoutBase.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithoutBase.Album)

	assert.Equal(t, 2, len(configWithoutBase.Album), "Expected 2 albums to be added")
	assert.Contains(t, configWithoutBase.Album, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithoutBase.Album, "ALBUM_2", "ALBUM_2 should be present")

	// configWithBaseOnly
	configWithBaseOnly := New()
	configWithBaseOnly.Album = []string{"BASE_ALBUM_1", "BASE_ALBUM_2"}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	err = configWithBaseOnly.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithBaseOnly.Album)

	assert.Equal(t, 2, len(configWithBaseOnly.Album), "Base albums should persist")
	assert.Contains(t, configWithBaseOnly.Album, "BASE_ALBUM_1", "BASE_ALBUM_1 should be present")
	assert.Contains(t, configWithBaseOnly.Album, "BASE_ALBUM_2", "BASE_ALBUM_2 should be present")
}

func TestAlbumAndPerson(t *testing.T) {
	testCases := []struct {
		name           string
		inputAlbum     []string
		inputPerson    []string
		expectedAlbum  []string
		expectedPerson []string
	}{
		{
			name:           "No empty values",
			inputAlbum:     []string{"album1", "album2"},
			inputPerson:    []string{"person1", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in album",
			inputAlbum:     []string{"album1", "", "album2", ""},
			inputPerson:    []string{"person1", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in person",
			inputAlbum:     []string{"album1", "album2"},
			inputPerson:    []string{"", "person1", "", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in both",
			inputAlbum:     []string{"", "album1", "", "album2"},
			inputPerson:    []string{"person1", "", "", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "All empty values",
			inputAlbum:     []string{"", "", ""},
			inputPerson:    []string{"", "", ""},
			expectedAlbum:  []string{},
			expectedPerson: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				Album:  tc.inputAlbum,
				Person: tc.inputPerson,
			}

			c.checkAssetBuckets()

			assert.Equal(t, tc.expectedAlbum, c.Album, "Album mismatch")
			assert.Equal(t, tc.expectedPerson, c.Person, "Person mismatch")
		})
	}
}

func TestCheckWeatherLocations(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "All fields present",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", Lon: "456", API: "abc123"},
				},
			},
			expected: "",
		},
		{
			name: "Missing name",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Lat: "123", Lon: "456", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: name",
		},
		{
			name: "Missing latitude",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lon: "456", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: latitude",
		},
		{
			name: "Missing longitude",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: longitude",
		},
		{
			name: "Missing API key",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", Lon: "456"},
				},
			},
			expected: "Weather location is missing required fields: API key",
		},
		{
			name: "Multiple missing fields",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City"},
				},
			},
			expected: "Weather location is missing required fields: latitude, longitude, API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			tt.config.checkWeatherLocations()

			output := strings.TrimSpace(buf.String())
			if tt.expected == "" {
				assert.Empty(t, output)
			} else {
				assert.NotEmpty(t, output)
			}
		})
	}
}
