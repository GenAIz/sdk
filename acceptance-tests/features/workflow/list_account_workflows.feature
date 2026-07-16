Feature: list account workflows
  To be able to list workflows of a solution on an account
  As an authenticated user
  I need to be able to create a solution, create a Smart Function, add a workflow node, login into an account, publish the solution and list the workflows from it

  Scenario: create solution with workflow
    Given the following parameters
      | folder      | oem             | solutionName | solutionDesc    | version | workflowHandle | workflowName | workflowDesc         |
      | my-solution | com.genaiz.test | My Solution  | A test solution | 1.0.0   | workflow-1     | workflow 1   | Workflow Description |
    When I run the command "sn create <folder> --oem=<oem> --name='<solutionName>' --description='<solutionDesc>' --workflow-handle=<workflowHandle> --workflow-name='<workflowName>' --workflow-desc='<workflowDesc>'"
    Then I should have a solution under "<folder>" named "<solutionName>" with oem "<oem>", handle "<folder>", description "<solutionDesc>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create smart function for workflow
    Given the scenario "create solution with workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle         | oem             | type     | version |
      | my-solution | bash-example | my-workflow-sf | com.genaiz.test | function | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: add smart function by folder
    Given the scenario "create smart function for workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeFolder     | nodeHandle          | sfHandle       | oem             | version |
      | my-solution | workflow-1     | my-workflow-sf | my-workflow-sf-node | my-workflow-sf | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeFolder>"
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

  Scenario: build smart function for workflow
    Given the scenario "add smart function by folder" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                     | handle         | oem             |
      | my-solution/my-workflow-sf | my-workflow-sf | com.genaiz.test |
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: publish workflow solution
    Given the scenario "login to broker" ran with condition "service_completed_successfully"
    And the scenario "build smart function for workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionOem     | functionHandle |
      | my-solution | com.genaiz.test | my-workflow-sf |
    And the workdir changes to "<folder>"
    When I run the command "sn publish"
    Then I should have a published solution "<folder>" with a smart function "<functionHandle>", oem "<functionOem>"

  Scenario: list solution workflows invalid fqdn
    Given the scenario "publish workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | solutionOem     | workflowHandle | workflowNodeCount |
      | my-solution | com.genaiz.test | my-solution    | 1                 |
    And the workdir changes to "<folder>"
    When I run the command "wf list <solutionOem>/<workflowHandle>"
    Then I should have an error "Error: invalid fqdn"

  Scenario: list solution workflows
    Given the scenario "publish workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | solutionOem     | workflowHandle | workflowVersion | workflowNodeCount |
      | my-solution | com.genaiz.test | my-solution    | 1.0.0           | 1                 |
    And the workdir changes to "<folder>"
    When I run the command "wf list <solutionOem>/<workflowHandle>:<workflowVersion>"
    Then I should have a list of workflows with workflow "<workflowHandle>" and a count of nodes equal to "<workflowNodeCount>"

  Scenario: list solution workflows with details
    Given the scenario "publish workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | solutionOem     | workflowHandle | nodeHandle          | sfOem           | sfHandle       | sfVersion |
      | my-solution | com.genaiz.test | workflow-1     | my-workflow-sf-node | com.genaiz.test | my-workflow-sf | 1.0.0     |
    And the workdir changes to "<folder>"
    When I run the command "wf list <solutionOem> --json"
    Then I should have a list of workflows with workflow "<workflowHandle>" with a node "<nodeHandle>" and a smart function oem "<sfOem>", handle "<sfHandle>" and version "<sfVersion>"
