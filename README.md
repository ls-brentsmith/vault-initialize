# vault-initialize

Tool to Initialize and Unseal a Vault Server

## Integration testing

### start a local vault server and run init against it

TODO: codify this

```bash
vault server -config=config.hcl

go run cmd/main.go
```

### docker image version

``` bash
# start vault server
vault server -config=config.hcl

# in a separate terminal
docker run -e VAULT_ADDR="http://host.docker.internal:8200" --rm -it gcr.io/ls-docker/shared/vault-initialize:0.1.0
```
