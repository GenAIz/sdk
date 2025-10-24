Feature: function run with the bash example
  To be able to run the bash example function
  As a local user
  I should be able to create, build and run the bash example

  Scenario: create bash example
    Given the following parameters
      | path              | oem            | version |
      | run-bash-function | com.genaiz.dev | 0.1.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example out of context
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | repository                       | version |
      | run-bash-function | com.genaiz.dev/run-bash-function | latest  |
    When I run the command "sf build --context=<path>"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: list bash example in context
    Given the scenario "build bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | repository                       | version |
      | run-bash-function | com.genaiz.dev/run-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf list"
    Then I should have a list with an image named "<repository>" with a version "<version>" and a docker image id

  Scenario: run bash example bad context
    Given the scenario "build bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | context      |
      | not-existing |
    When I run the command "sf run --context=<context>"
    Then I should have an error for field "sf.build.context"

  Scenario: run bash example bad file
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | file         |
      | run-bash-function | not-existing |
    When I run the command "sf run --file=<file>"
    Then I should have an error for field "sf.build.file"

  Scenario: run bash example bad tag
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | tag         |
      | run-bash-function | n--t.valid. |
    And the workdir changes to "<path>"
    When I run the command "sf run --tag=<tag>"
    Then I should have an error for field "sf.build.tag"

  Scenario: run bash example bad mount-in
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | mountIn     |
      | run-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf run --mount-in=<mountIn>"
    Then I should have an error for field "sf.run.input"

  Scenario: run bash example bad mount-out
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | mountOut    |
      | run-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf run --mount-out=<mountOut>"
    Then I should have an error for field "sf.run.output"

  Scenario: run bash example bad mount-log
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | mountLog    |
      | run-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf run --mount-log=<mountLog>"
    Then I should have an error for field "sf.run.log"

  Scenario: run bash example bad mount-var
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | mountVar    |
      | run-bash-function | /_not_valid |
    And the workdir changes to "<path>"
    When I run the command "sf run --mount-var=<mountVar>"
    Then I should have an error for field "sf.run.var"

  Scenario: run bash example bad prefix
    Given the scenario "building bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | prefix       |
      | run-bash-function | -bad-prefix- |
    And the workdir changes to "<path>"
    When I run the command "sf run --prefix=<prefix>"
    Then I should have an error for field "sf.run.prefix"

  Scenario: run bash example bad image
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | image                 |
      | run-bash-function | does-not-exist:latest |
    And the workdir changes to "<path>"
    When I run the command "sf run --image=<image>"
    Then I should have an error "image not found"

  Scenario: run bash example
    Given the scenario "build bash example out of context" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | oem            | handle            | version |
      | run-bash-function | com.genaiz.dev | run-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf run"
    Then I should have a build confirmation with repository "<oem>/<handle>:<version>" and a docker image id
    And I should have a run confirmation with repository "<oem>/<handle>:<version>" and a docker container id
