package trade

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/AlexSugak/getsky-trade/src/board"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HTTPServer holds http server info
type HTTPServer struct {
	binding      string
	board        board.Board
	log          logrus.FieldLogger
	httpListener *http.Server
	quit         chan os.Signal
	done         chan struct{}
}

// NewHTTPServer creates new http server
func NewHTTPServer(binding string, board board.Board, log logrus.FieldLogger) *HTTPServer {
	return &HTTPServer{
		binding: binding,
		board:   board,
		log: log.WithFields(logrus.Fields{
			"prefix": "trade.http",
		}),
		quit: make(chan os.Signal, 1),
		done: make(chan struct{}),
	}
}

// Run starts http listener and returns error if any
func (s *HTTPServer) Run() error {
	log := s.log
	log.Infof("HTTP service start at %s", s.binding)
	defer log.Info("HTTP service stop")
	signal.Notify(s.quit, os.Interrupt)

	r := s.setupRouter()

	s.httpListener = &http.Server{
		Addr:         s.binding,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      r,
	}

	errorC := make(chan error)
	go func() {
		if err := s.httpListener.ListenAndServe(); err != nil {
			log.Error(err)
			errorC <- err
		}
	}()

	select {
	case err := <-errorC:
		return err
	case <-s.quit:
		return s.Shutdown()
	}
}

// Shutdown shuts down the http listener
func (s *HTTPServer) Shutdown() error {
	s.log.Info("HTTP service shutting down")
	close(s.done)

	// Create a deadline to wait for.
	wait := time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	return s.httpListener.Shutdown(ctx)
}

func (s *HTTPServer) setupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api", ErrorHandler(s, APIInfoHandler(s))).Methods("GET")
	r.HandleFunc("/api/adverts/sell/latest", ErrorHandler(s, LatestSellAdvertsHandler(s))).Methods("GET")
	r.HandleFunc("/api/adverts/buy/latest", ErrorHandler(s, LatestBuyAdvertsHandler(s))).Methods("GET")

	return r
}

// APIHandler is a custom hadler function used internally to define api endpoint handlers
type APIHandler func(w http.ResponseWriter, r *http.Request) error

// ErrorHandler wraps APIHandler and converts it to http.Handler by handling any returned error
func ErrorHandler(s *HTTPServer, h APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			// TODO: define custom error type to store HTTP error code
			s.log.Errorf("Error in handler - %s", err)
			http.Error(w, err.Error(), 500)
		}
	}
}

// APIInfoResponse holds basic API information
type APIInfoResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     int    `json:"version"`
}

// APIInfoHandler returns api info
// Method: GET
// Content-type: application/json
// URI: /api
func APIInfoHandler(s *HTTPServer) APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		info := APIInfoResponse{
			Name:        "trade API",
			Description: "trade API provides endpoints to enable posting and searching Skycoin buy and sell adverts",
			Version:     1,
		}

		return json.NewEncoder(w).Encode(info)
	}
}

// LatestSellAdvertsHandler returns 10 latest sell adverts
// Method: GET
// Content-type: application/json
// URI: /api/adverts/sell/latest
func LatestSellAdvertsHandler(s *HTTPServer) APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		adverts, err := s.board.GetLatestAdverts(board.Sell, 10)
		if err != nil {
			return err
		}

		return json.NewEncoder(w).Encode(adverts)
	}
}

// LatestBuyAdvertsHandler returns 10 latest buy adverts
// Method: GET
// Content-type: application/json
// URI: /api/adverts/buy/latest
func LatestBuyAdvertsHandler(s *HTTPServer) APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		adverts, err := s.board.GetLatestAdverts(board.Buy, 10)
		if err != nil {
			return err
		}

		return json.NewEncoder(w).Encode(adverts)
	}
}
