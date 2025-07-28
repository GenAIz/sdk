Feature: function publish
  To be able to publish a function
  As an authenticated user
  I should be able to create, build and publish the bash example

  Scenario: create bash example
    Given the following parameters
      | recipe       | configType | fqdn           | handle          | oem        | name            | type     | version |
      | bash-example | yaml       | dev.genaiz.com | my-bash-example | com.genaiz | My Bash Example | function | 0.0.1   |
    When I run the command "sf create --recipe=<recipe> --configType=<configType> --fqdn=<fqdn> --handle=<handle> --oem=<oem> --name=<name> --type=<type> --version=<version> <handle>"
    Then I should have a function named "<name>"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | handle          | oem        |
      | my-bash-example | com.genaiz |
    When I run the command "sf build -c=<handle>"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: login bash example
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=success"
    When I run the command "ac login <orchestrator> --username=_bash_example@genaiz.com"
    Then I should have a session id

  Scenario: publish bash example
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the scenario "login bash example" ran with condition "service_completed_successfully"
    And the registry is running with condition: "service_healthy"
    And a provisioned function with token
    And the following parameters
      | handle          | oem        | version |
      | my-bash-example | com.genaiz | 0.1     |
    When I run the command "sf publish <orchestrator> --version=<version>"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>-dev"