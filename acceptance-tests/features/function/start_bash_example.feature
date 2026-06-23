Feature: function start with the bash example
  To be able to start the bash example function
  As a local user
  I should be able to create, build and start the bash example

  Scenario: create bash example
    Given the following parameters
      | path                | oem            | version |
      | start-bash-function | com.genaiz.test | 1.0.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example in context
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | repository                         | version |
      | start-bash-function | com.genaiz.test/start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf build"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: list bash example in context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | repository                         | version |
      | start-bash-function | com.genaiz.test/start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: start bash example bad context
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | context      |
      | not-existing |
    When I run the command "sf start --context=<context>"
    Then I should have an error for field "function.build.context"

  Scenario: start bash example bad file
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | file         |
      | start-bash-function | not-existing |
    When I run the command "sf start --file=<file>"
    Then I should have an error for field "function.build.file"

  Scenario: start bash example bad repository
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | repository         |
      | start-bash-function | n--t.valid. |
    And the workdir changes to "<path>"
    When I run the command "sf start --repository=<repository>"
    Then I should have an error for field "function.build.repository"

  Scenario: start bash example bad mount-in
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | mountIn     |
      | start-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf start --mount-in=<mountIn>"
    Then I should have an error for field "function.test.input"

  Scenario: start bash example bad mount-out
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | mountOut    |
      | start-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf start --mount-out=<mountOut>"
    Then I should have an error for field "function.test.output"

  Scenario: start bash example bad mount-log
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | mountLog    |
      | start-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf start --mount-log=<mountLog>"
    Then I should have an error for field "function.test.log"

  Scenario: start bash example bad mount-var
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | mountVar    |
      | start-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf start --mount-var=<mountVar>"
    Then I should have an error for field "function.test.var"

  Scenario: start bash example bad prefix
    Given the scenario "building bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | prefix       |
      | start-bash-function | -bad-prefix- |
    And the workdir changes to "<path>"
    When I run the command "sf start --prefix=<prefix>"
    Then I should have an error for field "function.test.prefix"

  Scenario: start bash example bad image
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | image                 |
      | start-bash-function | does-not-exist:latest |
    And the workdir changes to "<path>"
    When I run the command "sf start --image=<image>"
    Then I should have an error "image not found"

  Scenario: start bash example preserving
    Given the scenario "build bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | oem            | handle              | version |
      | start-bash-function | com.genaiz.test | start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf start --preserve"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a start confirmation with repository "<oem>/<handle>:<version>" and a docker container id

  Scenario: list bash example preserved
    Given the scenario "start bash example preserving" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | oem            | handle              | version |
      | start-bash-function | com.genaiz.test | start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with a container named "<oem>-<handle>-0" with an image "<oem>/<handle>:<version>" and a docker container id

  Scenario start bash example replacing preserving
    Given the scenario "list bash example preserved" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | oem            | handle              | version |
      | start-bash-function | com.genaiz.test | start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf start --preserve --replace"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a start confirmation with repository "<oem>/<handle>:<version>" and a docker container id

  Scenario: list bash example replaced
    Given the scenario "start bash example replacing preserving" ran with condition "service_completed_successfully"
    And the following parameters
      | path                | oem            | handle              | version |
      | start-bash-function | com.genaiz.test | start-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with a container named "<oem>-<handle>-0" with an image "<oem>/<handle>:<version>" and a docker container id
