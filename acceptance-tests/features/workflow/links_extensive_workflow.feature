Feature: function links for an extensive workflow
  To be able to add and remove links from a workflow
  As a developer
  I should be able to create a solution, create two functions, add workflow nodes and link their ports

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem             | handle     | version | workflowHandle | workflowName     | workflowDescription |
      | my-solution | com.genaiz.test | solution-1 | 0.1.1   | workflow-1     | Default Workflow | default workflow    |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --version=<version> --workflow-handle=<workflowHandle>"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create first function
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle         | oem             | version |
      | my-solution | bash-example | first-function | com.genaiz.test | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"

  Scenario: create second function
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle          | oem             | version |
      | my-solution | bash-example | second-function | com.genaiz.test | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe> --version=<version>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"

  Scenario: add first node
    Given the scenario "create first function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | nodeName   | nodeHandle          | oem             | version |
      | my-solution | workflow-1     | first-function | first-node | first-function-node | com.genaiz.test | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder> --name=<nodeName>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeName>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add output data port
    Given the scenario "create first function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                     | portHandle | portName  | portDesc      |
      | my-solution/first-function | test-port  | Test Port | A description |
    And the workdir changes to "<folder>"
    When I run the command "sf data output add <portHandle> --name='<portName>' --description='<portDesc>'"
    Then I should have an output port under "<folder>" named "<portName>" with handle "<portHandle>" and description "<portDesc>"

  Scenario: add second node
    Given the scenario "create second function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder  | nodeName    | nodeHandle           | oem             | version |
      | my-solution | workflow-1     | second-function | second-node | second-function-node | com.genaiz.test | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder> --name=<nodeName>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeName>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add input data port
    Given the scenario "create first function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                      | portHandle |
      | my-solution/second-function | test-port  |
    And the workdir changes to "<folder>"
    When I run the command "sf data input add <portHandle>"
    Then I should have an input port under "<folder>" named "<portHandle>" with handle "<portHandle>" and no description

  Scenario: link first and second nodes with invalid right port
    Given the scenario "add first node" ran with condition "service_completed_successfully"
    And the scenario "add second node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | firstNodeHandle     | firstPort | firstHandle  | secondNodeHandle     |
      | my-solution | workflow-1     | first-function-node | test-port | first-handle | second-function-node |
    And the workdir changes to "<folder>"
    When I run the command "wf links add <workflowHandle> <firstNodeHandle>[<firstPort>]:<secondNodeHandle>"
    Then I should have an error with "the right side of a link must have a data port"

  Scenario: link first and second nodes
    Given the scenario "add first node" ran with condition "service_completed_successfully"
    And the scenario "add second node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | firstNodeHandle     | firstPort | firstHandle  | secondNodeHandle     | secondPort |
      | my-solution | workflow-1     | first-function-node | test-port | first-handle | second-function-node | test-port  |
    And the workdir changes to "<folder>"
    When I run the command "wf links add <workflowHandle> <firstNodeHandle>[<firstPort>]:<secondNodeHandle>[<secondPort>]"
    Then I should have a link under "<folder>", workflow "<workflowHandle>" with left side handle "<firstNodeHandle>", port "<firstPort>" and a right side handle "<secondNodeHandle>" on port "<secondPort>"

  Scenario: remove link
    Given the scenario "link first and second nodes" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | firstNodeHandle     | functionPort | firstHandle  | secondNodeHandle     | secondPort |
      | my-solution | workflow-1     | first-function-node | test-port    | first-handle | second-function-node | test-port  |
    And the workdir changes to "<folder>"
    When I run the command "wf links rm <workflowHandle> <firstNodeHandle>[<firstPort>]:<secondNodeHandle>[<secondPort>]"
    Then I should have no links under "<folder>", workflow "<workflowHandle>"
