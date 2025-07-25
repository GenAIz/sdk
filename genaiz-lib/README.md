# GenAIz Library

## Makefile

Building the genaiz library independently of other modules can be useful to produce test and coverage reports:

```shell
cd genaiz-lib && make all
```

To view the coverage report on your default browser:

```shell
cd genaiz-lib && make coverage
```

For more information about all provided targets

```shell
cd genaiz-lib && make help
```

## Packages

### lang

Contains facilities to help condensed common language constructs. These may need to be replaced by judicious go libraries, but for now the goal was to keep the amount of dependencies to review low and so lang is nothing more than a Utils package.

### mock

Contains code that we need for executing unit tests. GO does not allow code to be shared across compilation units when it is enclosed within an X_test.go file. This means you either need to put the logic right next to the units you want to run coverage reports on, or in some place where the coverage reports do not matter.

Again this may be moved in the future, depending on the amount of improvements needed on the unit test metrics.