# GenAIz SDK Toolkits

## Modules

### [GenAIz SmartFunction Kit](genaiz/README.md)

Handles building, debugging, running, testing and publishing Smart Functions to a GenAIz Broker Platform.

### [GenAIz SmartFunction Integration Tool](genaiz-it/README.md)

Handles testing of service deployments with the GenAIz SmartFunction Kit.

### [Compose Services](compose/README.md)

The root [compose.yaml](compose.yaml) includes a list of profiles from the compose module, which can be used to test GenAIz Kits with the various services they require. This typically mean a Broker, a Container Server (containerd), and a Docker Registry.

All of these need to be accessible for the various kit functionality to run.

> [!CAUTION]
> Most of the tooling provided should eventually be embedded inside the genaiz-it kit and be available for testing any deployments of the Broker and Registry services.

### [Wiremock Mappings](wiremock/README.md)

A series of utility compose targets which can be used to test the various commands of the GenAIz SmartFunction Kit.

## Development Guide

### Building All Modules

Individual modules with make files will answer their own targets, but most will have *clean, build* and *test* at a minimum.

#### [GenAIz Makefile](genaiz/README.md#makefile)

#### [GenAIz OAuth](genaiz-oauth/README.md)

The oauth utility kit provides functionality to:

- Generate Certificate Authority ECDSA key pair files
- Generate Signed Certificate ECDSA key pair files
- Manage JWKS public key stores
- Create and Decode JWT Signed tokens

These facilities should be used to integrate token authentication on  [CNCF Distribution Registry](https://distribution.github.io/distribution/spec/auth/jwt/). With source code available on [GitHub](https://github.com/distribution/distribution)

#### [GenAIz-IT Makefile](genaiz-it/README.md#makefile)

### Testing GenAIz

## Hacking Guide

### Mocked Docker Registry

#### Compose

The profile mock-registry can be used from the root [compose.yaml](compose.yaml) file. It creates a registry bound to the host port 5000 with username and password written to `out/.registry/auth`

```shell
docker compose --profile mock-registry up
docker login -u genaiz_user -p genaiz_pass
```

Username and password can be modified from the [.env](.env) file. Also, one may change the user id and group id used to write to the host folder with the following:

```shell
DOCKER_UID=1010 DOCKER_GID=1010 docker compose --profile mock-registry up
```

#### Manually

If the need for testing docker login commands with the sdk arises, we can deploy a local registry featuring a basic username/password login preamble:

```shell
mkdir -p .registry && cd .registry
mkdir -p auth
docker run --entrypoint htpasswd httpd:2 -Bbn genaiz_user genaiz_pass > auth/htpasswd
docker run -d -p 5000:5000 --restart=always --name registry_genaiz  \
  -v `pwd`/auth:/auth  \
  -e "REGISTRY_AUTH=htpasswd"  \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm"  \
  -e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" registry:latest
docker login -u genaiz_user -p genaiz_pass
cd -
```