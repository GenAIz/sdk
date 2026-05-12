Feature: solution publish with simple solution
  To be able to publish the simple solution
  As an authenticated user
  I should be able to create a solution, create a smart function, build the smart function, add a workflow node and publish the solution

  Scenario: create simple solution
    Given the following parameters
      | folder      | oem            | handle     | name        | version | workflowDesc         | workflowHandle | workflowName |
      | my-solution | com.genaiz.dev | solution-1 | My Solution | 0.1.1   | workflow description | workflow-1     | workflow one |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --name='<name>' --version=<version> --workflow-desc='<workflowDesc>' --workflow-handle=<workflowHandle> --workflow-name='<workflowName>'"
    Then I should have a solution under "<folder>" named "<name>" with oem "<oem>", handle "<handle>", description "<name>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create simple solution function
    Given the scenario "create simple solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | recipe       | oem            | type     | version |
      | my-solution | my-function | bash-example | com.genaiz.dev | function | 0.1.1   |
    And the workdir changes to "<solution>"
    When I run the command "sf create <folder> --recipe=<recipe>"
    Then I should have a function under "<folder>" named "<folder>" with handle "<folder>", oem "<oem>", version "<version>" and type "<type>"

  Scenario: build simple solution function
    Given the scenario "create simple solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | handle     | oem            |
      | my-solution | my-function | function-1 | com.genaiz.dev |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: add workflow node to simple solution
    Given the scenario "build simple solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | workflowHandle | handle | name        | description    | functionOem    | functionHandle | functionVersion |
      | my-solution | workflow-1     | node-1 | Single Node | My Single Node | com.genaiz.dev | function-1     | 0.1.1           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <handle> --name='<name>' --description='<description>' --sf=<functionOem>/<functionHandle>:<functionVersion>"
    Then I should have a workflow node under "<solution>" with handle "<handle>", name "<name>", description "<description>" and smart function "<functionOem>/<functionHandle>:<functionVersion>"

  Scenario: publish simple solution no session
    Given the scenario "build simple solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    |
      | my-solution |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have an the error "not logged in"

  Scenario: login simple solution
    Given the scenario "publish simple solution no session" ran with condition "service_completed_successfully"
    And the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish simple solution
    Given the scenario "login simple solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | handle     | functionHandle |
      | my-solution | solution-1 | function-1     |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have a published solution "<handle>" with a smart function "<functionHandle>"
