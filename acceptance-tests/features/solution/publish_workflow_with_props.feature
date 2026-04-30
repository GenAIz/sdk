Feature: workflow publish with node properties
  To be able to publish a workflow with node properties
  As an authenticated user
  I should be able to create a solution, create a smart function, add a property specification, add it as a node on a workflow, build the smart function and publish the solution

  Scenario: create workflow solution
    Given the following parameters
      | folder      | oem            | handle     | name        | version | workflowDesc     | workflowHandle | workflowName     |
      | my-solution | com.genaiz.dev | solution-1 | My Solution | 0.1.1   | default workflow | default        | Default Workflow |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --name='<name>' --version=<version>
    Then I should have a solution under "<folder>" named "<name>" with oem "<oem>", handle "<handle>", description "<name>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create workflow solution function
    Given the scenario "create workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | recipe       | oem            | type     | version |
      | my-solution | my-function | bash-example | com.genaiz.dev | function | 0.1.1   |
    And the workdir changes to "<solution>"
    When I run the command "sf create <folder> --recipe=<recipe>"
    Then I should have a function under "<folder>" named "<folder>" with handle "<folder>", oem "<oem>", version "<version>" and type "<type>"

  Scenario: add property to function
    Given the scenario "create workflow solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                  | key      | type   |
      | my-solution/my-function | TEST_KEY | STRING |
    And the workdir changes to "<folder>"
    When I run the command "sf prop add <key>"
    Then I should have a property specification for key "<key>" under path "<path>" with type "<type>"

  Scenario: add workflow node to workflow solution
    Given the scenario "add property to function" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | workflowHandle | function    | handle           | functionOem    | functionVersion |
      | my-solution | default        | my-function | my-function-node | com.genaiz.dev | 0.1.1           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <function>"
    Then I should have a workflow node under "<solution>" with handle "<handle>", name "<name>" and smart function "<functionOem>/<function>:<functionVersion>"

  Scenario: add property to workflow node
    Given the scenario "add workflow node to workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflow | function    | node             | key      | value       |
      | my-solution | default  | my-function | my-function-node | TEST_KEY | value_added |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <workflow> <function> <key> '<value>'"
    Then I should have a property on workflow node "<node>" under workflow "<workflow>" with key "<key>" and value "<value>"

  Scenario: build workflow solution function
    Given the scenario "add property to workflow node" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | handle      | oem            |
      | my-solution | my-function | my-function | com.genaiz.dev |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: create orchestrator session
    Given the orchestrator is running with condition: "service_healthy"
    And the scenario "build workflow solution function" ran with condition "service_completed_successfully"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish workflow solution
    Given the scenario "create orchestrator session" ran with condition: "service_healthy"
    And the following parameters
      | solution    | functionHandle |
      | my-solution | my-function    |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have a published solution "<solution>" with a smart function "<functionHandle>"
