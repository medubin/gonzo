package server

import (
	"fmt"
)

type Server struct {
	// mux *http.ServeMux
}

type Request struct {
	Hi string `json:"hi"`
}

type Response struct {
	Bye string `json:"bye"`
}

func ServeIt(r Request) (Response, error) {
	return Response{
		Bye: fmt.Sprintf("%s:bye", r.Hi),
	}, nil
}

func NewServer() *Server {
	return &Server{}
}

// func (s *Server) Task(w http.ResponseWriter, r *http.Request) {
// 	println(r.URL.String())
// 	// println(r.Body.Read("hi"))
// 	js, _ := json.Marshal("test")
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(js)
// }

// func (s *Server) Task() string {
// 	return "test"
// }

// func Initialize(server Server) {
// 	server.mux = http.NewServeMux()
// 	// server.mux.HandleFunc("/task/")
//   // mux.HandleFunc("/tag/", server.Tag)
//   // mux.HandleFunc("/due/", server.Due)
// }

// func (s *Server) handle(endpoint string, handler func(any) (any, error))  {
// 	s.mux.HandleFunc(endpoint, func(http.ResponseWriter, *http.Request) {
// 		resp, err :=
// 	})
// 	// mux.HandleFunc("/tag/", server.Tag)
// }
