Feature: function publish with the bash example
  To be able to publish the bash example function
  As an authenticated user
  I should be able to create, build and publish the bash example

  Scenario: create bash example
    Given the following parameters
      | recipe       | configType | handle          | oem            | name            | type     | version |
      | bash-example | yaml       | my-bash-example | com.genaiz.dev | My Bash Example | function | 0.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --config-type=<configType> --handle=<handle> --oem=<oem> --name='<name>' --type=<type> --version=<version>"
    Then I should have a function under "<handle>" named "<name>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: publish bash example bad oem
    Given the following parameters
      | handle          | oem       | version |
      | my-bash-example | --invalid | 0.1.2   |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "sf.publish.oem"

  Scenario: publish bash example bad handle
    Given the following parameters
      | handle    | oem            | version |
      | --invalid | com.genaiz.dev | 0.1.2   |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --handle=<handle> --version=<version>"
    Then I should have an error for field "sf.publish.handle"

  Scenario: publish bash example bad version
    Given the following parameters
      | handle          | oem            | version |
      | my-bash-example | com.genaiz.dev | 00.0.1  |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --version=<version>"
    Then I should have an error for field "sf.publish.version"

  Scenario: publish bash example bad name
    Given the following parameters
      | handle          | oem            | name                                                                                                                                                                                                                                                                          |
      | my-bash-example | com.genaiz.dev | a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long a string too long |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --name='<name>'"
    Then I should have an error for field "sf.publish.name"

  Scenario: publish bash example bad type
    Given the following parameters
      | handle          | oem            | version | type    |
      | my-bash-example | com.genaiz.dev | 0.1.2   | invalid |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --version=<version> --type<type>"
    Then I should have an error for field "sf.publish.type"

  Scenario: publish bash example bad arch
    Given the following parameters
      | handle          | oem            | version | arch    |
      | my-bash-example | com.genaiz.dev | 0.1.2   | invalid |
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --version=<version> --arch=<arch>"
    Then I should have an error for field "sf.publish.arches"

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
      | handle          | oem        | version |
      | my-bash-example | com.genaiz | 0.1.2   |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish --version=<version>"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>-rc-0"
    And the config "Genaiz.yaml" should have "sf.publish.version" set to <version>
    And I should get an output with "<oem>/<handle>, version <version> to <orchestrator>"

  Scenario: publish bash example rebuild no update
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the scenario "login bash example" ran with condition "service_completed_successfully"
    And the registry is running with condition: "service_healthy"
    And the following parameters
      | handle          | oem        | previous | version |
      | my-bash-example | com.genaiz | 0.1.2    | 0.1.3   |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish --version=<version> --rebuild --no-update"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>-rc-0"
    And the config "Genaiz.yaml" should have "sf.publish.version" set to <previous>
    And I should get an output with "<oem>/<handle>, version <version> to <orchestrator>"

  Scenario: publish bash example duplicate aborting
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the scenario "login bash example" ran with condition "service_completed_successfully"
    And the registry is running with condition: "service_healthy"
    And the following parameters
      | handle          | oem        | previous | version |
      | my-bash-example | com.genaiz | 0.1.2    | 0.1.3   |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish --broker=<orchestrator> --context=<handle> --version=<version> --rebuild"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>-rc-0"
    And the config "Genaiz.yaml" should have "sf.publish.version" set to <previous>
    And I should get an output with "<orchestrator>/<oem>/<handle>:<version>-rc-0"
