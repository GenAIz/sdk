# GenAIz OAuth Utility Toolkit

## Makefile

Building the oauth tool with its associated make file can install the application, its manual pages and associated resources. To install locally:

```shell
cd genaiz-oauth && make all
```

To build a docker image of the sdk

```shell
cd genaiz-oauth && make docker
```

To execute tests alone or to get coverage metrics

```shell
cd genaiz-oauth && make coverage
```

For more information about all provided targets

```shell
cd genaiz-oauth && make help
```

## Commands

### certs

The certs command can be used to generate and/or parse certificates necessary for signing or parsing OAUTH tokens. 

> [!CAUTION]
> A certificate Authority (CA) is always necessary to sign the public key (CERT) used to sign a token. This CA certificate can be provided without generating a self-signed one.

```shell
./genaiz-oauth certs --help
```

### jwks

The jwks command is a utility for crafting jwks files, which are used by the CNCF Distribution registry for storing the public keys of signers and other deployments.

```shell
./genaiz-oauth jwks --help
```

### tokens

The tokens command can be used to generate and/or decode signed tokens using the certificates built by the [certs](#certs) command

```shell
./genaiz-oauth tokens --help
```

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

