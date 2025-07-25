# GenAIz SDK

## Modules

### [Genaiz SmartFunction Toolkit](genaiz/README.md)

Handles building, debugging, running, testing and publishing Smart Functions to a GenAIz Broker Platform.

### [GenAIz Integration Test Toolkit](genaiz-it/README.md)

Handles testing of service deployments with the GenAIz SmartFunction Kit. It offers the following functionality:

- CNCF Distribution Registry bootstrapping for tests
- Wiremock deployment to simulate what the SDK expects out of an orchestrator deployment
- Cucumber Genaiz Feature definitions and their associated step definitions

### [GenAIz Library](genaiz-lib/README.md)

The library of common functionality used by other GenAIz go modules. Providing:

- Common language enhancements
- Common support code for unit tests as GO unit tests can not share code across compilation units

### [GenAIz OAuth Utility Toolkit](genaiz-oauth/README.md)

The oauth utility kit provides functionality to:

- Generate Certificate Authority ECDSA key pair files
- Generate Signed Certificate ECDSA key pair files
- Manage JWKS public key stores
- Create and Decode JWT Signed tokens

These facilities should be used to integrate token authentication on  [CNCF Distribution Registry](https://distribution.github.io/distribution/spec/auth/jwt/). With source code available on [GitHub](https://github.com/distribution/distribution)

## Development Guide

### Building All Modules

Individual modules with make files will answer their own targets, but most will have *clean, build* and *test* at a minimum.

#### [GenAIz Makefile](genaiz/README.md#makefile)

#### [GenAIz-IT Makefile](genaiz-it/README.md#makefile)

#### [GenAIz-Lib Makefile](genaiz-lib/README.md#makefile)

#### [GenAIz-OAuth Makefile](genaiz-oauth/README.md#makefile)

### Testing GenAIz

TODO