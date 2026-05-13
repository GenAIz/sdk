Feature: function test with the bash example
  To be able to test the bash example function
  As a local user
  I should be able to create, build and test the bash example

  Scenario: create bash example
    Given the following parameters
      | path               | oem            | version |
      | test-bash-function | com.genaiz.dev | 1.0.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                        | version |
      | test-bash-function | com.genaiz.dev/test-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf build"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: list bash example in context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                        | version |
      | test-bash-function | com.genaiz.dev/test-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: test bash example bad context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | context      |
      | not-existing |
    When I run the command "sf test --context=<context>"
    Then I should have an error for field "function.build.context"

  Scenario: test bash example bad file
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | file         |
      | test-bash-function | not-existing |
    When I run the command "sf test --file=<file>"
    Then I should have an error for field "function.build.file"

  Scenario: test bash example bad repository
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository         |
      | test-bash-function | n--t.valid. |
    And the workdir changes to "<path>"
    When I run the command "sf test --repository=<repository>"
    Then I should have an error for field "function.build.repository"

  Scenario: test bash example bad mount-in
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | mountIn     |
      | test-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf test --mount-in=<mountIn>"
    Then I should have an error for field "function.test.input"

  Scenario: test bash example bad mount-out
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | mountOut    |
      | test-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf test --mount-out=<mountOut>"
    Then I should have an error for field "function.test.output"

  Scenario: test bash example bad mount-log
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | mountLog    |
      | test-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf test --mount-log=<mountLog>"
    Then I should have an error for field "function.test.log"

  Scenario: test bash example bad mount-var
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | mountVar    |
      | test-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf test --mount-var=<mountVar>"
    Then I should have an error for field "function.test.var"

  Scenario: test bash example bad prefix
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | prefix       |
      | test-bash-function | -bad-prefix- |
    And the workdir changes to "<path>"
    When I run the command "sf test --prefix=<prefix>"
    Then I should have an error for field "function.test.prefix"

  Scenario: test bash example bad image
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | image                 |
      | test-bash-function | does-not-exist:latest |
    And the workdir changes to "<path>"
    When I run the command "sf test --image=<image>"
    Then I should have an error "image not found"

  Scenario: test bash example
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | oem            | handle             | version |
      | test-bash-function | com.genaiz.dev | test-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf test"
    Then I should have a test session attached with the test output
