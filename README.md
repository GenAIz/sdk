# GenAIz CLI

<sub>Genaiz Version 1.0.0</sub>

The GenAIz CLI is a tool for creating, building and publishing Smart Functions to the GenAIz Orchestration platform. It
also provides toolkits to manage Orchestrated Workspaces and execute their Workflows.

* [Installation](#installation)
* [Running the GenAIz CLI](#running-the-genaiz-cli)
    * [Quickstart Examples](#quickstart-examples)
    * [Result Values](#result-values)
    * [Auto-Completion](#auto-completion)
    * [Account Management](#account-management)
    * [Workspace Management](#workspace-management)
    * [Credentials Management](#credentials-management)
* [Development Guide](#development-guide)
    * [Prerequisites](#prerequisites)
    * [Building from source](#building-from-source)
* [Troubleshooting](#troubleshooting)
    * [GoLang](#golang)
    * [Docker](#docker)
* [Contributing](#contributing)
* [License](#license)

## Installation

The `genaiz` command line tool can be downloaded from the [Releases](https://github.com/GenAIz/sdk/releases) of this
repository. There is also a [GenAIz Setup Action](https://www.github.com/GenAIz/genaiz-setup-action), which can be used
to configure GitHub Action Workflows.

## Running the GenAIz CLI

Working with the GenAIz CLI is a simple series of commands. You may run `genaiz --help` at any time to get a list of
available commands and their options.

### Quickstart Examples

#### Creating a simple solution

The following example creates a solution **solution-1** and a smart function **my-bash-example** using the recipe
**bash-example**. It then builds the function with the default version assigned by create,
[1.0.0](acceptance-tests/docs/function/index.md#version), assigns a single node to the **default** workflow,
authenticates a user with [dev.genaiz.com](acceptance-tests/docs/account/index.md#host) and publishes the solution to
the broker.

```bash
genaiz sn create mySolutionDir --name="My Solution" --handle="solution-1" \
  --oem="com.genaiz.examples" --description="A Description"
cd mySolutionDir
genaiz sf create mySmartFunction --recipe="bash-example" \
  --oem="com.genaiz.examples" --name="My Bash Example"
cd mySmartFunction
genaiz sf build
cd ..
genaiz wf nodes add default node1 --name="Single Node" \
  --description="My Single Node" \
  --sf="com.genaiz.examples/mySmartFunction:1.0.0"
genaiz ac login dev.genaiz.com
genaiz sn publish
```

#### Running a simple function

The following example creates a smart function **function-1**, without a parent solution. It then proceeds to build it
out of context. We can then list and run the smart function on a local **Docker** installation within the folder created
or out of it.

```bash
genaiz sf create function-1 --oem="com.genaiz.examples" --recipe="bash-example"
genaiz sf build --context=function-1/
cd function-1
genaiz sf list
genaiz sf run
```

> [!NOTE]
> Run disposes of any container created for the operation

#### Testing a simple function with Properties

The following example creates a smart function **function-2**, without a parent solution. We then proceed to add
property specifications to the function. The example completes with building the function in context, and testing it
with the console attached using `test`.

```bash
genaiz sf create function-2 --oem="com.genaiz.examples" \
  --recipe="bash-example"
cd function-2
genaiz sf prop add MY_KEY --name="Key Example" \
  --default-value=10 --type=int
genaiz sf prop env MY_KEY 12
genaiz sf build
genaiz sf test
```

#### Testing a simple function with Data Link Properties

The following example creates a smart function **function-3**, without a parent solution. We then proceed to create and
publish a data link **datalink-1** and add it to the smart function for provisioning on a broker.

```bash
genaiz sf create function-3 --oem="com.genaiz.examples" --version="0.2.0" \
  --recipe="bash-example"
cd function-3
genaiz dk create datalink-1 --oem="com.genaiz.examples" --version="1.0.0" \
  --description="My DataLink"
cd function-3
genaiz data source add com.genaiz/datalink-1:1.0.0
genaiz sf build
genaiz ac login dev.genaiz.com
genaiz sf publish
```

### Result Values

Support for result values is provided on `genaiz sf publish` only. From manually edited settings, for example:

```yaml
function:
  build:
    repository: com.genaiz/result-values
    version: latest
  publish:
    handle: result-values
    name: result-values
    oem: com.genaiz
    resultvalues:
      - excel
      - valid-set
    type: function
    version: 1.0.0
  run:
    input: run/in
    log: run/{timestamp}/log
    output: run/{timestamp}/out
    var: run/{timestamp}/var
```

### Auto-Completion

Auto-completion in Bash and other shell types can be achieved by generating the corresponding script file and sourcing
it inside the terminal session:

```bash
gennaiz completion --help
genaiz completion bash > ~/.config/genaiz/completion.sh
source ~/.config/genaiz/completion.sh
```

For Bash, the source instruction can be added to `$HOME/.bashrc` or `$HOME/.bash_profile`, depending on your terminal
shell's configuration.

### Account Management

#### login to OIDC enabled brokers

The account command can be used to authenticate with multiple OIDC enabled brokers with valid account credentials.

```bash
genaiz ac login dev.genaiz.com
genaiz ac list
genaiz ac login lab.genaiz.com
genaiz ac activate dev.genaiz.com
genaiz ac list --json
```

The command should print a JSON array with the sessions currently known to the SDK. Logging out of a session can be
achieved by specifying the host account to log out or by simply logging out of the active session without adding any
host argument:

```bash
genaiz ac logout dev.genaiz.com
genaiz ac logout
```

#### Environment logins for CI/CD

To facilitate a secure in-memory mechanism for handling credentials, the genaiz cli can use a combination of environment
variables.

* `GENAIZ_AUTH_URL`: the url that would be used on genaiz ac login command
* `GENAIZ_AUTH_SESSION`: the session token emitted by a previous ac login command for that host

These 2 bits of information can be used on CI/CD configurations to avoid having to write a `.auth` file under
`$HOME/.cache/genaiz` for a given session. The session token should typically be held in the **Secrets** of the CI/CD
pipeline in need of calling the genaiz cli.

### Workspace Management

#### Creating a Workspace

### Credentials Management

The GenAIz CLI provides a way to manage Data Sources and Data Stores which would be used either with a Workspace on an
Orchestration, or locally with the run, start and test commands.

The facilities are there to fortify usage of potentially sensitive information on local developer environments.
Currently, those keys are still held in clear text in .env files, which is not recommended at large.

#### Creating a Data Locker

For Sources (read-only) or Stores (read/write) connections, a **locker** can be created and updated with:

```bash
genaiz locker init
genaiz locker source add myLocalHandle com.genaiz/my-link:1.0.0
genaiz locker source update myLocalHandle MyKey MyValue
gpg --decrypt myKeyfile.gpg | genaiz locker source update myLocalHandle MySecretKey
```

#### Orchestration with Data Lockers

The locker can be used to create a DataSource on the same account used to define the Datalink. Lockers can only ever be
used in the context of a Workspace to avoid orphaned Active definitions to remain on server.

```bash
genaiz workspace data source link my-workspace \
  my-workflow-handle my-node-handle myLocalHandle
```

And removed from the workspace flow later:

```bash
genaiz workspace data source unlink my-workspace \
  my-workflow-handle myLocalHandle
```

#### Local runs with Data Lockers

Locally, the locker should be used when invoking the Smart Function run, start and test commands:

```bash
genaiz sf run --env-locked=myLocalHandle --locker=myFilePath
```

## Development Guide

### Prerequisites

To be able to build the project you will need to install the following tools for your platform. You can find more
information at the links provided here:

* [Golang](https://go.dev/doc/install) - All tools and the SDK are written in Go
* [GNU Make](https://www.gnu.org/software/make/manual/make.html) - Not strictly required, but all build scripts are
  driven through Makefiles
* [Docker](https://docs.docker.com/engine/install/) - The local execution and build runtime requires a local Docker
  installation

### Building from Source

```bash
cd genaiz-cli
make genaiz/install
```

> [!TIP]
> It is possible $HOME/go/bin is not included within a users' path. That can be fixed with
> `export PATH=$PATH:$HOME/go/bin` added to the users' .bashrc, .bash_profile or .profile file, depending on the
> OS and distribution.

Building the individual modules can be achieved in the same manner or simply by running the install target on the root
project:

```bash
make install
```

Print a summary of all make targets by entering

```bash
make help
```

## Troubleshooting

### Golang

#### Go tools tag

We use 2 different environment builds. The `prod` build is the build configured by default under the `Makefile` of the
project. It builds with an HTTP request gate which denies connections on `http` addresses. When this is not practical,
for testing reasons, the `dev` tag can be enabled to exempt `localhost` from this restriction.

When using an IDE to run go tests, you will have to configure it manually with `-tags dev`. Under **Intellij**, this is
configured under

* `Run/Debug Configurations`
    * `Edit configuration templates`
        * `Go Test`
            * `Go tool arguments`

#### Tests are failing with `permission denied`

Golang's testing tools rely on executing code from the `/tmp` folder. On many modern system /tmp is now mounted to
[tmpfs](https://www.kernel.org/doc/html/latest/filesystems/tmpfs.html) with `noexec`. It is a recommendation
from [CIS](https://www.cisecurity.org/). You may have to fix GO's toolchain environment adding:

```bash
export GOTMPDIR="/var/tmp"
```

to your `$HOME/.bashrc` or `$HOME/.bash_profile` file. Note that `/var/tmp` is just another one of those "locations"
that is usually the target for temporary files while compiling projects. You could very well just use `$HOME/go/tmp`,
depending on how your partition layout was created, if `/var/tmp` is also "secured".

#### Tests are failing with nil panic traces

We use a flavor of **Monkey Patching** to capture calls to various core methods like os.Exit and fmt.Printf. These
require the source to be compiled without function inlining enabled:

`-gcflags=-l`

The option is set in the `Makefile` requiring it for the `test` and `report` targets, but if you use an IDE Run
configuration you will have to configure it manually. Under **Intellij**, this is configured under

* `Run/Debug Configurations`
    * `Edit configuration templates`
        * `Go Test`
            * `Go tool arguments`

### Docker

#### Connection to /var/run/docker.sock failing

Check that docker is running the following under Systemd:

```bash
systemctl status docker
```

Under OpenRC:

```bash
rc-service docker status
```

Check that the permissions on /var/run/docker.sock allows your user to connect to it:

```bash
ls -l /var/run/docker.sock
groups
```

Your user should be in the same group owning the docker.sock file.

If all conditions are met, and you are using a Linx/amd64 distributions, write to `sdk 'at' genaiz.com` with the
following details:

* Docker version
* OS and version
* If you can build with docker build
* the output of the command with --log-level=debug

## Contributing

To contribute to this project, please follow these steps:

1. Fork this repository to your personal account.
2. Create a new branch for your feature or bug fix.
3. Commit and push changes to your forked repository.
4. Open a Pull Request (PR) to allow merging into the upstream repository.

## License

The GenAIz CLI is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.

&copy; 2018 - 2026 GenAIz. All rights reserved.