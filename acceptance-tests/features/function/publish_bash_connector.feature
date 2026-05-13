Feature: connector publish with the bash example
  To be able to publish the bash example connector
  As an authenticated user
  I should be able to create the function, add data sources, data stores and outbound proxies, build and publish the connector

  Scenario: create bash connector no version
    Given the following parameters
      | recipe       | handle            | oem            | type      | version |
      | bash-example | my-bash-connector | com.genaiz.dev | connector | 1.0.0   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --type=<type>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: login bash connector
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: create data link for data source
    Given the scenario "login bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | handle         | oem            | version |
      | my-bash-connector | datalink-src-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<folder>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: create data link for data store
    Given the scenario "create data link for data source" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | handle         | oem            | version |
      | my-bash-connector | datalink-str-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<folder>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data source to connector
    Given the scenario "create data link for data source" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | handle         | oem            | version |
      | my-bash-connector | datalink-src-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add data store to connector
    Given the scenario "add data source to connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | handle         | oem            | version |
      | my-bash-connector | datalink-str-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data store add <handle>:<version> --oem=<oem>"
    Then I should have a data store under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add outbound proxy to connector
    Given the scenario "add data store to connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | hostValue      | hostPort | options | flags       |
      | my-bash-connector | dev.genaiz.com | 22       | --tcp   | active, tcp |
    And the workdir changes to "<folder>"
    When I run the command "sf data proxy add <hostValue>:<hostPort> <options>"
    Then I should have an outbound proxy under "<folder>" with host "<hostValue>", port "<hostPort>" and flags "<flags>"

  Scenario: build bash connector
    Given the scenario "add outbound proxy to connector" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | folder            | oem        |
      | my-bash-connector | com.genaiz |
    And the workdir changes to "<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: publish bash connector
    Given the scenario "build bash connector" ran with condition "service_completed_successfully"
    And the scenario "login bash example" ran with condition "service_completed_successfully"
    And the registry is running with condition: "service_healthy"
    And the following parameters
      | handle            | oem        | version |
      | my-bash-connector | com.genaiz | 1.0.0   |
    And the execution group "<docker_gid>"
    And the workdir changes to "<handle>"
    When I run the command "sf publish"
    Then I should have a docker image tagged "registry/<oem>/<handle>:<version>-rc-0"
    And the config "Genaiz.yaml" should have "function.publish.version" set to <version>
    And I should get an output with "<oem>/<handle>, version <version> to <orchestrator>"
