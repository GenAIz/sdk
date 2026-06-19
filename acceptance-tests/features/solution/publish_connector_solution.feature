Feature: solution publish with a simple connector
  To be able to publish a solution with a simple connector
  As an authenticated user
  I should be able to create a solution, create a smart function connector, build the smart function, add a workflow node and publish the solution

  Scenario: create simple solution with defaults
    Given the following parameters
      | folder      | oem            | version | workflowDesc     | workflowHandle | workflowName     |
      | my-solution | com.genaiz.dev | 1.0.0   | default workflow | default        | Default Workflow |
    When I run the command "sn create <folder> --oem=<oem>"
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create bash connector with version
    Given the following parameters
      | folder      | recipe       | handle       | oem            | type      | version |
      | my-solution | bash-example | my-connector | com.genaiz.dev | connector | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --type=<type> --version=<version>"
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
      | folder       | handle       | oem            | version | name       |
      | my-connector | data-input-1 | com.genaiz.dev | 0.2.0   | Data Input |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem> --version=<version> --name='<name>'"
    Then I should have a datalink under "<folder>" named "<name>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data source to connector
    Given the scenario "create data link for data source" ran with condition "service_completed_successfully"
    And the following parameters
      | folder       | handle       | oem            | version |
      | my-connector | data-input-1 | com.genaiz.dev | 0.2.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: create data link for data store
    Given the scenario "create data link for data source" ran with condition "service_completed_successfully"
    And the following parameters
      | folder       | handle        | oem            | version |
      | my-connector | data-output-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<folder>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data store to connector
    Given the scenario "add data source to connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder       | handle        | oem            | version |
      | my-connector | data-output-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data store add <handle>:<version> --oem=<oem>"
    Then I should have a data store under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add outbound proxy to connector
    Given the scenario "add data store to connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder       | hostValue      | hostPort | options | flags       |
      | my-connector | dev.genaiz.com | 22       | --tcp   | active, tcp |
    And the workdir changes to "<folder>"
    When I run the command "sf data proxy add <hostValue>:<hostPort> <options>"
    Then I should have an outbound proxy under "<folder>" with host "<hostValue>", port "<hostPort>" and flags "<flags>"

  Scenario: build bash connector
    Given the scenario "add outbound proxy to connector" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | folder       | oem        |
      | my-connector | com.genaiz |
    And the workdir changes to "<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: add workflow node to simple solution
    Given the scenario "build bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | workflowHandle | handle | description | functionOem    | functionHandle | functionVersion |
      | my-solution | default        | node-1 |             | com.genaiz.dev | my-connector   | 0.1.1           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <handle> --sf=<functionOem>/<functionHandle>:<functionVersion>"
    Then I should have a workflow node under "<solution>" with handle "<handle>", oem "<oem>", description "<description>" and smart function "<functionOem>/<functionHandle>:<functionVersion>"

  Scenario: publish simple solution
    Given the scenario "add workflow node to simple solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | functionHandle |
      | my-solution | my-connector   |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have a published solution "<handle>" with a smart function "<functionHandle>"
