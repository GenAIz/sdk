Feature: function build with the bash example
  To be able to build the bash example function
  As a developer
  I should be able to create and build the bash example

  Scenario: create bash example
    Given the following parameters
      | recipe       | configType | handle          | oem            | name            | type     | version |
      | bash-example | yaml       | my-bash-example | com.genaiz.dev | My Bash Example | function | 0.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --config-type=<configType> --handle=<handle> --oem=<oem> --name='<name>' --type=<type> --version=<version>"
    Then I should have a function under "<handle>" named "<name>" with version "<version>"

  Scenario: build bash example bad context
    Given the following parameters
      | context     | file       | tag                |
      | bad-context | Dockerfile | com.genaiz.dev/tag |
    When I run the command "sf build --context=<context> --file=<file> --tag=<tag>"
    Then I should have an error for field "sf.build.context"

  Scenario: build bash example bad file
    Given the following parameters
      | context      | file          | tag                |
      | bash-example | badDockerfile | com.genaiz.dev/tag |
    When I run the command "sf build --context=<context> --file=<file> --tag=<tag>"
    Then I should have an error for field "sf.build.file"

  Scenario: build bash example bad tag
    Given the following parameters
      | context      | file       | tag                     |
      | bash-example | Dockerfile | com.genaiz.dev/bad..tag |
    When I run the command "sf build --context=<context> --file=<file> --tag=<tag>"
    Then I should have an error for field "sf.build.tag"

  Scenario: build bash example no context file
    Given the following parameters
      | tag                |
      | com.genaiz.dev/tag |
    When I run the command "sf build --tag=<tag>"
    Then I should have an error for field "sf.build.file"

  Scenario: build bash example no arguments
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | handle          | oem            |
      | my-bash-example | com.genaiz.dev |
    And the workdir changes to "<handle>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"
