tfstate-http-backend-vault
==========================

[Terraform](https://www.terraform.io) [HTTP state backend](https://www.terraform.io/docs/backends/types/http.html) implementation using [Vault](https://www.vaultproject.io/) to store state.

Usage
-----

Install:

```bash
go get github.com/ngrow/tfstate-http-backend-vault
```

Run:

```
VAULT_TOKEN=... tfstate-http-backend-vault
```

(you can use ~/.vault-token file written by `vault login` command)

Add terraform backend:

```
terraform {
  backend "http" {
    address = "http://localhost:8080/"
  }
}
```

TODO
----

* Lock support
