package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-systemd/activation"
	"github.com/hashicorp/go-hclog"
	vault_api "github.com/hashicorp/vault/api"
)

var (
	flagAddr      string
	flagVaultAddr string
)

func init() {
	flag.StringVar(&flagAddr, "listen-addr", ":8080", "Address to serve")
	flag.StringVar(&flagVaultAddr, "vault-addr", "", "Adress of Vault server")
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func newVaultClient() (*vault_api.Client, error) {
	var config *vault_api.Config
	if flagVaultAddr != "" {
		config = &vault_api.Config{
			Address: flagVaultAddr,
		}
	}
	vault_client, err := vault_api.NewClient(config)
	if err != nil {
		hclog.Default().Error("failed to create Vault API client", "err", err)
		return nil, err
	}

	if _, ok := os.LookupEnv("VAULT_TOKEN"); !ok {
		token_filename := os.Getenv("HOME") + "/.vault-token"
		if !fileExists(token_filename) {
			var err = errors.New("there is no such file: " + token_filename)
			hclog.Default().Error("vault token is not specified", "err", err)
			return nil, err
		}
		token, err := ioutil.ReadFile(token_filename)
		if err != nil {
			hclog.Default().Error("couldn't read token", "err", err)
			return nil, err
		}
		//hclog.Default().Info("couldn't read token", "token", string(token))
		vault_client.SetToken(strings.TrimSpace(string(token)))
	}

	if res, err := vault_client.Sys().Health(); err != nil {
		hclog.Default().Error("failed to check vault's health", "err", err)
		return nil, err
	} else {
		hclog.Default().Info("vault's healthy", "version", res.Version)
	}

	return vault_client, nil
}

func getListener(flagAddr string) (net.Listener, error) {
	if listeners, err := activation.Listeners(); err != nil {
		return net.Listen("tcp", flagAddr)
	} else {
		return listeners[0], nil
	}
}

func main() {
	flag.Parse()

	hclog.Default().SetLevel(hclog.Info)

	secret_path := os.Getenv("TFSTATE_SECRET_PATH")
	if secret_path == "" {
		secret_path = "secret/data/tfstate"
	}
	hclog.Default().Info("tfstate stored at", "path", secret_path)

	vault_client, err := newVaultClient()
	if err != nil {
		os.Exit(1)
	}
	vault := vault_client.Logical()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				hclog.Default().Info("couldn't read request body", "err", err)
				return
			}
			encoded_body := base64.StdEncoding.EncodeToString(body)
			_, err = vault.Write(secret_path, map[string]interface{}{"data": map[string]interface{}{"data": encoded_body}})
			if err != nil {
				w.WriteHeader(500)
				hclog.Default().Info("couldn't write tfstate", "err", err)
				return
			}
			hclog.Default().Info("wrote state", "data", string(body))
		case "GET":
			secret, err := vault.Read(secret_path)
			if err != nil {
				w.WriteHeader(500)
				hclog.Default().Info("couldn't read tfstate", "err", err)
				return
			}
			if secret == nil {
				return
			}
			data := secret.Data["data"]
			if data != nil {
				data = (data.(map[string]interface{}))["data"]
			}
			if data != nil {
				decoded_data, err := base64.StdEncoding.DecodeString(data.(string))
				if err != nil {
					hclog.Default().Info("couldn't decode base64", "err", err)
					w.WriteHeader(500)
					return
				}
				_, err = w.Write(decoded_data)
				if err != nil {
					hclog.Default().Info("error writing response", "err", err)
					return
				}
				hclog.Default().Debug("read state", "data", string(decoded_data))
			}
		case "LOCK":
			w.WriteHeader(405)
		case "UNLOCK":
			w.WriteHeader(405)
		default:
			w.WriteHeader(405)
		}
	})

	lis, err := getListener(flagAddr)
	if err != nil {
		hclog.Default().Error("failed to get listener", "err", err)
		os.Exit(1)
	}

	if err := http.Serve(lis, nil); err != nil {
		hclog.Default().Error("failed to serve HTTP", "err", err)
	}
}
