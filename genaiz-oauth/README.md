# GenAIz OAuth Utility Kit

## Examples

### Simple Scenario

```shell
cd genaiz-oauth
go build
# Generate CA Cert, a Server Cert and a rootCertBundle or trust store
./genaiz-oauth certs generate --genBundle \
  --svCert=out/oauth/server.cert \
  --svKey=out/oauth/server.key \
  --svCN=dev.genaiz.com \
  out/oauth
# Create a JWKS store file containing all the public keys
./genaiz-oauth jwks create --jwks=out/oauth/keys.jwks out/oauth/server.cert
# Create a token to a token.txt file
./genaiz-oauth tokens generate \
  --signCert=out/oauth/server.cert \
  --signKey=out/oauth/server.key \
  --aud=test.genaiz.com \
  --op=pull \
  --repo=com.genaiz/myapp \
  --exp=60 \
  --out=out/oauth/token.txt \
  out/oauth
# Decode the generated token
./genaiz-oauth tokens decode --in=out/oauth/token.txt out/oauth/
```

