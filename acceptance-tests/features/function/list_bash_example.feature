Feature: function list with the bash example
  To be able to list the bash example function
  As a local user
  I should be able to create, build, run and list the bash example

  Scenario: create bash example
    Given the following parameters
      | path               | oem             | version |
      | list-bash-function | com.genaiz.test | 1.0.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                         | version |
      | list-bash-function | com.genaiz.test/list-bash-function | latest  |
    When I run the command "sf build"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: start bash example out of context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                         | version |
      | list-bash-function | com.genaiz.test/list-bash-function | latest  |
    When I run the command "sf start --context=<path> --preserve"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a start confirmation with repository "<oem>/<handle>:<version>" and a docker container id

  Scenario: list bash example out of context
    Given the scenario "run bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                         | version |
      | list-bash-function | com.genaiz.test/list-bash-function | latest  |
    When I run the command "sf list --context=<path>"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id
    And I should have a list with a container named "<oem>-<handle>-0" with an image "<oem>/<handle>:<version>" and a docker container id
