package samlidp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seriousben/dev-identity-provider/internal/storage"
	"github.com/zenazn/goji/web"
)

// HandleListUsers handles the `GET /users/` request and responds with a JSON formatted list
// of user names.
func (s *Server) HandleListUsers(c web.C, w http.ResponseWriter, r *http.Request) {
	users, err := s.Store.List("/users/")
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Users []string `json:"users"`
	}{Users: users})
}

// HandleGetUser handles the `GET /users/:id` request and responds with the user object in JSON
// format. The HashedPassword field is excluded.
func (s *Server) HandleGetUser(c web.C, w http.ResponseWriter, r *http.Request) {
	user := storage.User{}
	err := s.Store.Get(fmt.Sprintf("/users/%s", c.URLParams["id"]), &user)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// HandlePutUser handles the `PUT /users/:id` request. It accepts a JSON formatted user object in
// the request body and stores it. If the PlaintextPassword field is present then it is hashed
// and stored in HashedPassword. If the PlaintextPassword field is not present then
// HashedPassword retains it's stored value.
func (s *Server) HandlePutUser(c web.C, w http.ResponseWriter, r *http.Request) {
	user := storage.User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	user.ID = c.URLParams["id"]

	err := s.Store.Put(fmt.Sprintf("/users/%s", c.URLParams["id"]), &user)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteUser handles the `DELETE /users/:id` request.
func (s *Server) HandleDeleteUser(c web.C, w http.ResponseWriter, r *http.Request) {
	err := s.Store.Delete(fmt.Sprintf("/users/%s", c.URLParams["id"]))
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
