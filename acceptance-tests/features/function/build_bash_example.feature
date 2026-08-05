Feature: function build with the bash example
  To be able to build the bash example function
  As a developer
  I should be able to create and build the bash example

  Scenario: create bash example
    Given the following parameters
      | recipe       | configType | handle          | oem             | name            | type     | version |
      | bash-example | yaml       | my-bash-example | com.genaiz.test | My Bash Example | function | 0.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --config-type=<configType> --oem=<oem> --name='<name>' --type=<type> --version=<version>"
    Then I should have a function under "<handle>" named "<name>", with handle "<handle>", oem "<oem>", version "<version>" and type "<type>"

  Scenario: build bash example bad context
    Given the following parameters
      | context     | file       | repository           |
      | bad-context | Dockerfile | com.genaiz.test/repo |
    When I run the command "sf build --context=<context> --file=<file> --repository=<repository>"
    Then I should have an error for field "function.build.context"

  Scenario: build bash example bad file
    Given the following parameters
      | context      | file          | repository           |
      | bash-example | badDockerfile | com.genaiz.test/repo |
    When I run the command "sf build --context=<context> --file=<file> --repository=<repository>"
    Then I should have an error for field "function.build.file"

  Scenario: build bash example bad repository
    Given the following parameters
      | context      | file       | repository                |
      | bash-example | Dockerfile | com.genaiz.test/bad..repo |
    When I run the command "sf build --context=<context> --file=<file> --repository=<repository>"
    Then I should have an error for field "function.build.repository"

  Scenario: build bash example no context file
    Given the following parameters
      | repository           |
      | com.genaiz.test/repo |
    When I run the command "sf build --repository=<repository>"
    Then I should have an error for field "function.build.file"

  Scenario: build bash example no arguments
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | handle          | oem             |
      | my-bash-example | com.genaiz.test |
    And the workdir changes to "<handle>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: build bash example for platform
    Given the scenario "build bash example no arguments" ran with condition "service_completed_successfully"
    And the following parameters
      | handle          | oem             | platform    |
      | my-bash-example | com.genaiz.test | linux/arm64 |
    And the workdir changes to "<handle>"
    When I run the command "sf build --platform=<platform"
    Then I should have dock image tagged "<oem>/<handle>:latest"
    And I should have the platform "<platform>" documented in the image metadata

  Scenario: build bash example for platform with legacy builder
    Given the scenario "build bash example no arguments" ran with condition "service_completed_successfully"
    And the following parameters
      | handle          | oem             | platform    |
      | my-bash-example | com.genaiz.test | linux/arm64 |
    And the workdir changes to "<handle>"
    When I run the command "sf build --platform=<platform --legacy-builder"
    Then I should have dock image tagged "<oem>/<handle>:latest"
    And I should have the platform "<platform>" documented in the image metadata
