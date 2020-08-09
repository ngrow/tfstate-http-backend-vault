package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	vault_api "github.com/hashicorp/vault/api"
)

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {
	secret_path := os.Getenv("TFSTATE_SECRET_PATH")
	if secret_path == "" {
		secret_path = "secret/data/tfstate"
	}
	vault_client, err := vault_api.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := os.LookupEnv("VAULT_TOKEN"); !ok {
		token_filename := os.Getenv("HOME") + "/.vault-token"
		if fileExists(token_filename) {
			token, err := ioutil.ReadFile(token_filename)
			if err != nil {
				log.Fatal("couldn't read token: ", err)
			}
			vault_client.SetToken(string(token))
		} else {
			log.Fatal("vault token is not specified")
		}
	}
	vault := vault_client.Logical()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't read request body:", err)
				return
			}
			_, err = vault.Write(secret_path, map[string]interface{}{"data": map[string]interface{}{"data": body}})
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't write tfstate:", err)
				return
			}
			log.Print("wrote state")
		case "GET":
			secret, err := vault.Read(secret_path)
			if err != nil {
				w.WriteHeader(500)
				log.Print("couldn't read tfstate:", err)
				return
			}
			if secret == nil {
				return
			}
			data := secret.Data["data"]
			if data != nil {
				w.Write(data.([]byte))
			}
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
