Feature: start workspace flow
  To be able to start a workspace flow, creating a new run
  As an authenticated user
  I should be able to create a solution, create a smart function, add it to the solution's default workflow
  I should be able to login to a broker and publish the solution
  I should be able to create a workspace, create a workspace flow with the solution workflow, and list the nodes created
  I should be able to start a workspace run, using the workspace name and the workflow handle

  Scenario: create solution for workspace
    Given the following parameters
      | folder       | oem             | version | workflowHandle | workflowName | workflowDesc     |
      | run-solution | com.genaiz.test | 1.0.0   | default        | Default      | default workflow |
    When I run the command "sn create <folder> --oem=<oem>
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create function for solution
    Given the scenario "create solution for workspace" ran with condition "service_completed_successfully"
    And the following parameters
      | solution     | folder       | recipe       | oem             | type     | version |
      | run-solution | run-function | bash-example | com.genaiz.test | function | 1.0.0   |
    And the workdir changes to "<solution>"
    When I run the command "sf create <folder> --recipe=<recipe>"
    Then I should have a function under "<folder>" named "<folder>" with handle "<folder>", oem "<oem>", version "<version>" and type "<type>"

  Scenario: build function for solution
    Given the scenario "create function for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder       | handle       | oem             |
      | my-solution | run-function | run-function | com.genaiz.test |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: add function node to workflow
    Given the scenario "build function for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | function     | workflowHandle | handle | functionOem     | functionVersion |
      | my-solution | run-function | default        | node-1 | com.genaiz.test | 1.0.0           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <function>"
    Then I should have a workflow node under "<solution>" with handle "<handle>", name "<handle>", description "" and smart function "<functionOem>/<function>:<functionVersion>"

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
      | solution     | functionHandle |
      | run-solution | run-function   |
    And the workdir changes to "<solution>"
    When I run the command "sn publish"
    Then I should have a published solution "<solution>" with a smart function "<functionHandle>"

  Scenario: create workspace
    Given the scenario "login for create workspace flow" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 3     |
    When I run the command "ws create <name> --json"
    Then I should have a workspace with name "<name>", a created timestamp the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace flow
    Given the scenario "publish solution for workspace" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName  | oem             | solutionHandle | solutionVersion | wfHandle |
      | test_workspace | com.genaiz.test | run-solution   | 1.0.0           | default  |
    When I run the command "ws flow create <workspaceName> <oem>/<solutionHandle>:<solutionVersion> <wfHandle> --json"
    Then I should have a workspace flow for workflow "<wfHandle>" and solution "<oem>/<solutionHandle>:<solutionVersion>"

  Scenario: start workspace flow
    Given the scenario "create workspace flow" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName  | oem             | solutionHandle | solutionVersion | wfHandle | runName |
      | test_workspace | com.genaiz.test | run-solution   | 1.0.0           | default  | my-name |
    When I run the command "ws flow start <workspaceName> <wfHandle> --name=<runName> --json"
    Then I should have a workspace run for workflow "<wfHandle>" named "<runName>"
