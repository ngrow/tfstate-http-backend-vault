package main

import (
	"io/ioutil"
	"log"
	"net/http"

	vault_api "github.com/hashicorp/vault/api"
)

func checkAuth(r *http.Request, vault *vault_api.Logical) (bool, string) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false, "request is missing basic auth credentials"
	}
	secret, err := vault.Read("secret/tfstate/users/" + username)
	if err != nil {
		return false, "couldn't get user " + username + " info from vault: " + err.Error()
	}
	if secret.Data["password"].(string) != password {
		return false, "invalid password for " + username
	}
	return true, ""
}

func main() {
	vault_client, err := vault_api.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	vault := vault_client.Logical()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ok, message := checkAuth(r, vault)
		if !ok {
			w.WriteHeader(401)
			log.Print(message)
			return
		}
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't read request body:", err)
				return
			}
			_, err = vault.Write("secret/tfstate/data", map[string]interface{}{"data": body})
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't write tfstate:", err)
				return
			}
			log.Print("wrote state")
		case "GET":
			secret, err := vault.Read("secret/tfstate/data")
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't read tfstate:", err)
				return
			}
			w.Write(secret.Data["data"].([]byte))
			log.Print("read state")
		case "LOCK":
			w.WriteHeader(405)
		case "UNLOCK":
			w.WriteHeader(405)
		default:
			w.WriteHeader(405)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
