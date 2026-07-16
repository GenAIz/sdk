Feature: list local workflows
  To be able to list workflows of a solution
  As a developer
  I need to be able to create a solution, create a Smart Function, add a workflow node and list the workflows from it

  Scenario: create solution with workflow
    Given the following parameters
      | folder      | oem             | version | workflowHandle | workflowName | workflowDesc         |
      | my-solution | com.genaiz.test | 1.0.0   | workflow-1     | workflow 1   | Workflow Description |
    When I run the command "sn create <folder> --oem=<oem> --workflow-handle=<workflowHandle> --workflow-name='<workflowName>' --workflow-desc='<workflowDesc>'"
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

  Scenario: list solution workflows in workdir
    Given the scenario "add smart function by folder" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | workflowNodeCount |
      | my-solution | workflow-1     | 1                 |
    And the workdir changes to "<folder>"
    When I run the command "wf list"
    Then I should have a list of workflows with workflow "<workflowHandle>" and a count of nodes equal to "<workflowNodeCount>"

  Scenario: list solution workflows with details under folder
    Given the scenario "publish workflow solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | solutionOem     | workflowHandle | nodeHandle          | sfOem           | sfHandle       | sfVersion |
      | my-solution | com.genaiz.test | workflow-1     | my-workflow-sf-node | com.genaiz.test | my-workflow-sf | 1.0.0     |
    When I run the command "wf list <folder> --json"
    Then I should have a list of workflows with workflow "<workflowHandle>" with a node "<nodeHandle>" and a smart function oem "<sfOem>", handle "<sfHandle>" and version "<sfVersion>"
