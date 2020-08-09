tfstate-http-backend-vault
==========================

[https://www.terraform.io](Terraform) [https://www.terraform.io/docs/backends/types/http.html](HTTP state backend) implementation using [https://www.vaultproject.io/](Vault) to store state.

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
