package server

import "net/http"

type server struct {
	mux         *http.ServeMux
	middlewares []http.HandlerFunc
}

func New() *server {
	return &server{mux: http.NewServeMux()}
}

func (self *server) Use(next http.HandlerFunc) {
	self.middlewares = append(self.middlewares, next)
}

func (self *server) Handle(pattern string, handler http.HandlerFunc) {
	self.mux.Handle(pattern, handler)
}

func (self *server) start() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, middleware := range self.middlewares {
			middleware.ServeHTTP(w, r)
		}
		self.mux.ServeHTTP(w, r)
	})
}

func (self *server) ListenAndServe(address string) {
	http.ListenAndServe(address, self.start())
}
