package socketcmd

/*  Copyright 2017 Ryan Clarke

    This file is part of Socketcmd.

    Socketcmd is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Socketcmd is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Socketcmd.  If not, see <http://www.gnu.org/licenses/>
*/

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

/* A WrapperAPI extends an enclosed Wrapper with high-level remote API operations.
 */
type WrapperAPI interface {
	Wrapper
	/* Listen on the given address and serve the command endpoint at the given path.
	 */
	Listen(addr, path string) error
	/* Default Handler function for the WrapperAPI. This method may be used to integrate
	 * the WrapperAPI into an existing API or extend it with other endpoints. This endpoint
	 * expects to receive a command sequence as an array of strings in JSON format. If the
	 * first element is a valid socketcmd Header then it will be used to parse the response.
	 * Otherwise, the configured ParseFunc will be used to generate a header based on the
	 * given command sequence. The response will by sent back as a JSON array of strings.
	 */
	CommandEndpoint(http.ResponseWriter, *http.Request)
}

type wrapperAPI struct {
	*wrapper
	c Client
}

func (api *wrapperAPI) Listen(addr, path string) error {
	if path == "" {
		path = "/"
	}
	http.HandleFunc(path, api.CommandEndpoint)
	return http.ListenAndServe(addr, nil)
}

func (api *wrapperAPI) CommandEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse command sequence from request body
	body := []string{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handlerErr(w, err, http.StatusBadRequest)
		return
	}

	// Send command sequence to wrapped process and collect response
	resp, err := api.c.Send(body...)
	if err != nil {
		if err == ErrCommandForbidden {
			log.Printf("attempted forbidden command: %v\n", body)
		}
		handlerErr(w, err, http.StatusInternalServerError)
		return
	}

	// Encode response and send back to the client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handlerErr(w, err, http.StatusInternalServerError)
		return
	}
}

func handlerErr(w http.ResponseWriter, err error, status int) {
	// Log the error to the console, set the response header, and send error in response body
	if err != ErrCommandForbidden {
		log.Println(err)
	}
	w.WriteHeader(status)
	io.WriteString(w, err.Error()+"\n")
}
