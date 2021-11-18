storage "inmem" {}

disable_mlock = "true"

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = "true"
}

api_addr = "http://127.0.0.1:8200"
cluster_addr = "https://127.0.0.1:8201"
ui = true

seal "gcpckms" {
  project     = "vault-on-cloud-run-323521"
  region      = "global"
  key_ring    = "vault-server"
  crypto_key  = "seal"
}
