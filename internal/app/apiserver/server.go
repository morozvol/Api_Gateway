package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
)

const (
	ctxKeyRequestID ctxKey = iota
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type ctxKey int8

type server struct {
	router *mux.Router
	logger *logrus.Logger
	config *Config
}

func newServer(config *Config) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		config: config,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.HandleFunc("/login", s.handleRedirectAuthService()).Methods("POST")
	s.router.HandleFunc("/signup", s.handleRedirectAuthService()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc("/profile", s.handleRedirectUserService()).Methods("GET", "PUT")
	private.HandleFunc("/basket", s.handleRedirectBasketService()).Methods("GET", "PUT")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key := w.Header().Get("Authorization"); key == "" {
			s.respond(w, r, http.StatusUnauthorized, nil)
		}

		//TODO - декодируем Token проверяем роли
		// если токен не действителен или у пользователя нет нужных ролей прерываем запрос прерываем

		next.ServeHTTP(w, r)
	})
}

func (s *server) handleRedirectAuthService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if url, err := url.Parse(fmt.Sprintf("http://%s", s.config.AuthService_Addr)); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(w, r)
		}
	}
}

func (s *server) handleRedirectUserService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if url, err := url.Parse(fmt.Sprintf("http://%s", "url UserService")); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(w, r)
		}
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *server) handleRedirectBasketService() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if url, err := url.Parse(fmt.Sprintf("http://%s", "url BaketService")); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(w, r)
		}
	}
}
