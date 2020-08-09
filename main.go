package main

import (
	"io/ioutil"
	"log"
	"net/http"

	vault_api "github.com/hashicorp/vault/api"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(401)
			log.Print("request is missing basic auth credentials")
			return
		}
		vault_client, err := vault_api.NewClient(nil)
		if err != nil {
			w.WriteHeader(500)
			log.Fatal(err)
		}
		vault_client.SetToken(password)
		vault := vault_client.Logical()
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't read request body:", err)
				return
			}
			_, err = vault.Write("secret/tfstate", map[string]interface{}{"data": body})
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't write tfstate:", err)
				return
			}
			log.Print("wrote state")
		case "GET":
			secret, err := vault.Read("secret/tfstate")
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
