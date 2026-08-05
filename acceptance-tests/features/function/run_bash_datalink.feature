Feature: connector run with the bash example and a datalink
  To be able to run the bash example connector with a datalink
  As a local user
  I should be able to create and build the bash example, create a datalink, add it and run the bash example with an associated environment file

  Scenario: create bash connector
    Given the following parameters
      | path              | oem             | version | type      |
      | run-bash-datalink | com.genaiz.test | 1.0.1   | connector |
    When I run the command "sf create <path> --oem=<oem> --version=<version> --recipe=bash-example --type=<type>"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | repository                        | version |
      | run-bash-datalink | com.genaiz.test/run-bash-datalink | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf build"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: list bash example out of context
    Given the scenario "build bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | repository                        | version |
      | run-bash-datalink | com.genaiz.test/run-bash-datalink | latest  |
    When I run the command "sf list --context=<path>"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: create datalink for bash example
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem             | handle     | version |
      | run-bash-datalink | com.genaiz.test | datalink-1 | 1.0.0   |
    When I run the command "dk create <oem>/<handle>:<version> <path>"
    Then I should have a datalink under "<path>/Genaiz.yaml" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add prop spec for datalink
    Given the scenario "create datalink for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem             | handle     | version | key    | defaultValue | type   |
      | run-bash-datalink | com.genaiz.test | datalink-1 | 1.0.0   | MY_KEY | test value   | STRING |
    And the workdir changes to "<path>"
    When I run the command  "dk prop add <oem>/<handle>:<version> <key> --default-value='<defaultValue>' --user-defined=false"
    Then I should have a "<type>" property spec under "<path>/Genaiz.yaml", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

  Scenario: add secret spec for datalink
    Given the scenario "add prop spec for datalink" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem             | handle     | version | key        |
      | run-bash-datalink | com.genaiz.test | datalink-1 | 1.0.0   | SECRET_KEY |
    And the workdir changes to "<path>"
    When I run the command  "dk prop add <oem>/<handle>:<version> <key> --secret --user-defined=false"
    Then I should have a "<type>" secret property spec under "<path>/Genaiz.yaml", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>", name "<key>"

  Scenario: login bash datalink
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario publish datalink for bash example
    Given the scenario "add secret spec for datalink" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem             | handle     | version |
      | run-bash-datalink | com.genaiz.test | datalink-1 | 1.0.0   |
    When I run the command  "dk publish <oem>/<handle>:<version> <path>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: add data source for datalink
    Given the scenario "publish datalink for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | configFile                   | oem             | handle     | version |
      | run-bash-datalink | ~/.config/genaiz/Genaiz.yaml | com.genaiz.test | datalink-1 | 1.0.0   |
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<path>" with datalink "<oem>/<handle>:<version>"
    And I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: run bash example
    Given the scenario "build bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem             | handle            | version |
      | run-bash-function | com.genaiz.test | run-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf run --env=SECRET_KEY=test"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a run confirmation with repository "<oem>/<handle>:<version>" and a docker container id
