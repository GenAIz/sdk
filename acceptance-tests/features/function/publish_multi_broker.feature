Feature: function publish, using multiple brokers
  To be able to publish the bash example function to multiple brokers
  As an authenticated user,
  I should be able to create, build and publish the bash example function to the orchestrator
  I should be able to modify and publish the bash example to a second orchestrator
  I should be able to publish the modified version back to the original orchestration

  Scenario: create bash example
    Given the following parameters
      | recipe       | handle        | oem             | type     | version |
      | bash-example | multi-example | com.genaiz.test | function | 1.0.0   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | handle        | oem             |
      | multi-example | com.genaiz.test |
    And the workdir changes to "<handle>"
    When I run the command "sf build"
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
    Given the scenario "login bash example" ran with condition "service_completed_successfully"
    And the orchestrator is running with condition: "service_healthy"
    And the following parameters
      | handle        | oem             | version |
      | multi-example | com.genaiz.test | 1.0.0   |
    And the workdir changes to <handle>
    When I run the command "sf publish"
    Then I should have a docker image tagged "<orchestrator>/<oem>/<handle>:<version>-rc-0"

  Scenario: modify bash example
    Given the scenario "publish bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder        | src    |
      | multi-example | app.sh |
    And the workdir changes to "<folder>"
    # will need a way to specify another image with a different entry point, or we'll need sed to be bundled and use a different entrypoint
    When I run the command "sed -i 's/sleep 4s/sleep 5s/' <src>"
    Then I should have a modified "<folder>/<src>" file

  Scenario: login second orchestrator
    Given the second orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <secondOrchestrator> --username=<username>"
    Then I should have an active session id with host "<secondOrchestrator>" for username "<username>" under path "<path>"

  Scenario: publish bash example to second orchestrator
    Given the scenario "login bash example" ran with condition "service_completed_successfully"
    And the second orchestrator is running with condition: "service_healthy"
    And the following parameters
      | handle        | oem             | version |
      | multi-example | com.genaiz.test | 1.0.0   |
    And the workdir changes to <handle>
    When I run the command "sf publish --rebuild"
    Then I should have a docker image tagged "<secondOrchestrator>/<oem>/<handle>:<version>-rc-0"

  Scenario: activate orchestrator session
    Given the second orchestrator is running with condition: "service_healthy"
    And the following parameters
      | path                     |
      | $HOME/.cache/genaiz.auth |
    When I run the command "ac activate <orchestrator>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish modified bash example
    Given the scenario "activate orchestrator session" ran with condition "service_completed_successfully"
    And the orchestrator is running with condition: "service_healthy"
    And the following parameters
      | handle        | oem             | version |
      | multi-example | com.genaiz.test | 1.0.0   |
    And the workdir changes to <handle>
    When I run the command "sf publish"
    Then I should have a docker image tagged "<orchestrator>/<oem>/<handle>:<version>-rc-1"
