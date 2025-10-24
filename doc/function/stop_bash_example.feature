Feature: function stop with the bash example
  To be able to stop the bash example function
  As a local user
  I should be able to create, build and stop the bash example

  Scenario: create bash example
    Given the following parameters
      | path               | oem            | version |
      | stop-bash-function | com.genaiz.dev | 0.1.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                        | version |
      | stop-bash-function | com.genaiz.dev/stop-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf build"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: list bash example in context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                        | version |
      | stop-bash-function | com.genaiz.dev/stop-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: start bash example
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | oem            | handle             | version | name                |
      | stop-bash-function | com.genaiz.dev | stop-bash-function | latest  | stop-bash-container |
    And the workdir changes to "<path>"
    When I run the command "sf start --name=<name> --preserve"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a start confirmation with repository "<oem>/<handle>:<version>" and a docker container id

  Scenario stop bash example
    Given the scenario "start bash example" ran with condition "service_healthy"
    And the following parameters
      | path               | oem            | handle             | version | name                |
      | stop-bash-function | com.genaiz.dev | stop-bash-function | latest  | stop-bash-container |
    And the workdir changes to "<path>"
    When I run the command "sf stop --name=<name>"
    Then I should have a stop confirmation with repository "<oem>/<handle>:<version>" and docker container id
