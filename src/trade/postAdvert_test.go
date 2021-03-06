package trade

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	tradedb "github.com/skycoin/getsky.org/db"
	"github.com/skycoin/getsky.org/src/util/logger"
	"github.com/skycoin/getsky.org/src/util/test"
	"github.com/stretchr/testify/require"
	validator "gopkg.in/go-playground/validator.v9"
)

type ServerTimeMock struct{}

func (s ServerTimeMock) Now() time.Time {
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	t, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	return t
}

func setupPostAdvertTests() func() {
	insertSQL(fmt.Sprintf("INSERT INTO `%s`.`Users` (UserName, Email, PasswordHash, TimeOffset, CountryCode, StateCode, City, PostalCode, DistanceUnits, Currency, Status) VALUES ('bob', 'bob@bob.com', 'foo', 0, 'US', 'CA', 'Los Angeles', '', 'mi', 'USD', 1)", dbName))

	return func() {
		clearTables()
	}
}

func TestPostBuyAdverts(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           []byte
		url            string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "should validate content type",
			method:         "POST",
			contentType:    "application/xml",
			body:           []byte(``),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "Invalid content type, expected application/jsonInvalid json request body: EOF",
		},
		{
			name:           "should validate format of the request entity",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(``),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid json request body: EOF",
		},
		{
			name:           "should validate mandatory fields of the entity",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"" }`),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `[{"key":"Recaptcha","message":"is required"}]`,
		},
		{
			name:           "should validate recapcha code",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"invalid_code"  }`),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `[{"key":"recaptcha","message":"is not valid"}]`,
		},
		{
			name:           "return 404 if user does not exist",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"not_existing_user", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"pass" }`),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `the user with the userName: 'not_existing_user' doesn't exist`,
		},
		{
			name:           "should save advert and return saved entity with 200 status",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"pass" }`),
			url:            "/api/postings/buy",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"type":2,"author":"bob","tradeCashInPerson":true,"tradeCashByMail":true,"tradeMoneyOrderByMail":true,"tradeOther":true,"amountFrom":"12.2","amountTo":null,"fixedPrice":null,"percentageAdjustment":"0","currency":"EUR","additionalInfo":"","travelDistance":12,"travelDistanceUoM":"km","countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"","status":1,"createdAt":"2013-02-03T19:54:00Z","expiredAt":"2013-03-03T19:54:00Z"}`,
		},
	}

	teardownTests := setupPostAdvertTests()
	defer teardownTests()

	for _, tc := range tests {
		name := fmt.Sprintf("test case: BuyAdvertHandler %s", tc.name)
		req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(tc.body))
		req.Header.Add("Content-type", tc.contentType)

		require.NoError(t, err)

		sql := sqlx.NewDb(db, "mysql")
		s := tradedb.NewStorage(sql)
		w := httptest.NewRecorder()
		u := tradedb.NewUsers(sql)
		server := &HTTPServer{board: s, users: u, checkRecaptcha: FakeRecaptchaChecker, log: logger.InitLogger()}
		server.validate = validator.New()
		server.serverTime = ServerTimeMock{}
		handler := server.setupRouter(test.StubSecure)

		handler.ServeHTTP(w, req)
		require.Equal(t, tc.expectedStatus, w.Code, name)
		require.Equal(t, tc.expectedBody, strings.TrimSuffix(w.Body.String(), "\n"), name)
	}
}

func TestPostSellAdverts(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           []byte
		url            string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "should validate content type",
			method:         "POST",
			contentType:    "application/xml",
			body:           []byte(``),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "Invalid content type, expected application/jsonInvalid json request body: EOF",
		},
		{
			name:           "should validate format of the request entity",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(``),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid json request body: EOF",
		},
		{
			name:           "should validate mandatory fields of the entity",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"" }`),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `[{"key":"Recaptcha","message":"is required"}]`,
		},
		{
			name:           "should validate recapcha code",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":"0","currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"invalid_code"  }`),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `[{"key":"recaptcha","message":"is not valid"}]`,
		},
		{
			name:           "return 404 if user does not exist",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"not_existing_user", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"pass" }`),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `the user with the userName: 'not_existing_user' doesn't exist`,
		},
		{
			name:           "should save advert and return saved entity with 200 status",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":0,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"pass" }`),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"type":1,"author":"bob","tradeCashInPerson":true,"tradeCashByMail":true,"tradeMoneyOrderByMail":true,"tradeOther":true,"amountFrom":"12.2","amountTo":null,"fixedPrice":null,"percentageAdjustment":"0","currency":"EUR","additionalInfo":"","travelDistance":12,"travelDistanceUoM":"km","countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"","status":1,"createdAt":"2013-02-03T19:54:00Z","expiredAt":"2013-03-03T19:54:00Z"}`,
		},
		{
			name:           "should return 400 if a request doesn't have both percentageAdjustment and fixedPrice",
			method:         "POST",
			contentType:    "application/json",
			body:           []byte(`{"author":"bob", "tradeCashInPerson":true, "tradeCashByMail":true, "tradeMoneyOrderByMail":true, "tradeOther":true, "amountFrom":12.2, "amountTo": null, "percentageAdjustment":null,"currency":"EUR", "additionalInfo":"", "travelDistance":12, "travelDistanceUoM":"km", "countryCode":"GR","stateCode":null,"city":"Athens","postalCode":"", "recaptcha":"pass" }`),
			url:            "/api/postings/sell",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `percentageAdjustment or fixedPrice has to have a value`,
		},
	}

	teardownTests := setupPostAdvertTests()
	defer teardownTests()

	for _, tc := range tests {
		name := fmt.Sprintf("test case: SellAdvertHandler %s", tc.name)
		req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(tc.body))
		req.Header.Add("Content-type", tc.contentType)

		require.NoError(t, err)

		sql := sqlx.NewDb(db, "mysql")
		s := tradedb.NewStorage(sql)
		w := httptest.NewRecorder()
		u := tradedb.NewUsers(sql)
		server := &HTTPServer{board: s, users: u, checkRecaptcha: FakeRecaptchaChecker, log: logger.InitLogger()}
		server.validate = validator.New()
		server.serverTime = ServerTimeMock{}
		handler := server.setupRouter(test.StubSecure)

		handler.ServeHTTP(w, req)
		require.Equal(t, tc.expectedStatus, w.Code, name)
		require.Equal(t, tc.expectedBody, strings.TrimSuffix(w.Body.String(), "\n"), name)
	}
}
