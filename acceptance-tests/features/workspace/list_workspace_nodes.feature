Feature: list workspace nodes
  To list the nodes of a workspace flow
  As an authenticated user
  I should be able to create a solution, create a smart function, add it to the solution's default workflow
  I should be able to login to a broker and publish the solution
  I should be able to create a workspace, create a workspace flow with the solution workflow, and list the nodes created

  Scenario: create solution for workspace
    Given the following parameters
      | folder      | oem             | version | workflowHandle | workflowName | workflowDesc     |
      | my-solution | com.genaiz.test | 1.0.0   | workflow-1     | Default      | default workflow |
    When I run the command "sn create <folder> --oem=<oem> --workflow-handle='<workflowHandle>'
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create function for solution
    Given the scenario "create solution for workspace" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | recipe       | oem             | type     | version |
      | my-solution | my-function | bash-example | com.genaiz.test | function | 1.0.0   |
    And the workdir changes to "<solution>"
    When I run the command "sf create <folder> --recipe=<recipe>"
    Then I should have a function under "<folder>" named "<folder>" with handle "<folder>", oem "<oem>", version "<version>" and type "<type>"

  Scenario: build function for solution
    Given the scenario "create function for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | handle     | oem             |
      | my-solution | my-function | function-1 | com.genaiz.test |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: add function node to workflow
    Given the scenario "build function for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | function    | workflowHandle | handle | functionOem     | functionHandle | functionVersion |
      | my-solution | my-function | workflow-1     | node-1 | com.genaiz.test | my-function    | 1.0.0           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <function>"
    Then I should have a workflow node under "<solution>" with handle "<handle>", name "<handle>", description "" and smart function "<functionOem>/<functionHandle>:<functionVersion>"

  Scenario: login for create workspace flow
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish solution for workspace
    Given the scenario "login simple solution" ran with condition "service_completed_successfully"
    And the scenario "add function node to workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | handle     | functionHandle |
      | my-solution | solution-1 | function-1     |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have a published solution "<handle>" with a smart function "<functionHandle>"

  Scenario: create workspace
    Given the scenario "login for create workspace flow" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 3     |
    When I run the command "ws create <name> --json"
    Then I should have a workspace with name "<name>", a created timestamp the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace flow
    Given the scenario "list solutions for workspace" ran with condition "service_completed_successfully"
    And the scenario "publish solution for workspace" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName  | oem             | solutionHandle | solutionVersion | wfHandle   |
      | test_workspace | com.genaiz.test | my-solution    | 1.0.0           | workflow-1 |
    When I run the command "ws flow create <workspaceName> <oem>/<solutionHandle>:<solutionVersion> <wfHandle> --json"
    Then I should have a workspace flow for workflow "<workflow-1>" and solution "<oem>/<solutionHandle>:<solutionVersion>"

  Scenario list workspace nodes for flow
    Given the scenario "create workspace flow" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName  | wfHandle   | nodeHandle |
      | test_workspace | workflow-1 | node-1     |
    When I run the command "ws node list <workspaceName> <wfHandle> --json"
    Then I should have a list of nodes with a node named "<nodeHandle>" and handle "<nodeHandle>"
