Feature: list account solutions
  To be able to list solutions available to an account
  As an authenticated user
  I need to be able to create a solution, create a Smart Function, add a workflow node, login into an account, publish the solution and list it back

  Scenario: create simple solution with name and description
    Given the following parameters
      | folder      | oem            | version | solutionName | solutionDesc         | workflowName     | workflowHandle | workflowDesc     |
      | my-solution | com.genaiz.dev | 1.0.0   | My Solution  | Solution Description | Default Workflow | default        | default workflow |
    When I run the command "sn create <folder> --oem=<oem> --name='<solutionName>' --description='<solutionDesc>'"
    Then I should have a solution under "<folder>" named "<solutionName>" with oem "<oem>", handle "<folder>", description "<solutionDesc>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create bash example function
    Given the scenario "create simple solution with name and description" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem            | type     | version |
      | my-solution | bash-example | my-function | com.genaiz.dev | function | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: add bash example node
    Given the scenario "create bash example function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionFolder | workflowHandle | nodeHandle       | oem            | version |
      | my-solution | my-function    | default        | my-function-node | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: login to broker
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish simple solution
    Given the scenario "login to broker" ran with condition "service_completed_successfully"
    And the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionHandle |
      | my-solution | my-function    |
    And the workdir changes to "<folder>"
    When I run the command "sn publish"
    Then I should have a published solution "<handle>" with a smart function "<functionHandle>"

  Scenario: list account solutions
    Given the scenario "publish simple solution" ran with condition "service_completed_successfully"
    And the following parameters
      | oem            | handle      | version | local |
      | com.genaiz.dev | my-function | 1.0.0   | false |
    When I run the command "sn list <oem>/<handle>:<version> --account=<orchestrator>"
    Then I should have a solution list with solution "<oem>/<handle>:<version>" named "<solutionName>" and local flag set to "<local>"
