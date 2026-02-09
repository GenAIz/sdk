Feature: function publish, un-synchronized, with the bash example
  To be able to recover from an empty local repository
  As an authenticated user
  I should be able to build and publish the same function that is already known to a broker

  Scenario: create bash example
    Given the following parameters
      | recipe       | handle      | oem            | type     | version |
      | bash-example | my-function | com.genaiz.dev | function | 0.1.0   |
    When I run the command "sf create <handle> --recipe=<recipe> --handle=<handle> --oem=<oem>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | folder      | oem            |
      | my-function | com.genaiz.dev |
    And the workdir changes to "<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<folder>:latest"

  Scenario: login bash example
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish bash example
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the scenario "login bash example" ran with condition "service_completed_successfully"
    And the registry is running with condition: "service_healthy"
    And the following parameters
      | folder      | oem            | version    |
      | my-function | com.genaiz.dev | 0.1.0-rc-0 |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>"
    And I should get an output with "<oem>/<handle>, version <version> to <orchestrator>"

  Scenario delete bash example image
    Given the scenario "publish bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | com.genaiz.dev | version    |
      | my-function | oem            | 0.1.0-rc-0 |
    When I run the command "sf list |grep '<oem/<folder>:latest' |awk '{print $3}' |xargs docker image rm -f"
    Then I should not have an image tagged "registry/<oem>/<folder>:<version>" locally

  Scenario: rebuild bash example
    Given the scenario "delete bash example image" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | folder      | oem            |
      | my-function | com.genaiz.dev |
    And the workdir changes to "<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<folder>:latest"

  Scenario: re-publish bash example
    Given the scenario "rebuild bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | oem            | version    |
      | my-function | com.genaiz.dev | 0.1.0-rc-0 |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish --version=<version>"
    Then I should get an error "smart function digest is already known"
