Feature: function links for an extensive workflow
  To be able to add and remove links from a workflow
  As a developer
  I should be able to create a solution, create two functions, add workflow nodes and link their ports

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem            | handle     | version | workflowHandle |
      | my-solution | com.genaiz.dev | solution-1 | 0.1.1   | workflow-1     |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --version=<version --workflow-handle=<workflowHandle>"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowHandle>", handle "<workflowHandle>" with description "<workflowHandle>"

  Scenario: create first function
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle         | oem            | version |
      | my-solution | bash-example | first-function | com.genaiz.dev | 0.1.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: create second function
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle          | oem            | version |
      | my-solution | bash-example | second-function | com.genaiz.dev | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe> --version=<version>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add first node
    Given the scenario "create first function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | nodeName   | nodeHandle   | oem            | version |
      | my-solution | my-workflow    | first-function | first-node | first-handle | com.genaiz.dev | 0.1.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder> --name=<nodeName> --handle=<nodeHandle>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeName>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add second node
    Given the scenario "create second function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder  | nodeName    | nodeHandle           | oem            | version |
      | my-solution | my-workflow    | second-function | second-node | second-function-node | com.genaiz.dev | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder> --name=<nodeName>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeName>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: link first and second nodes
    Given the scenario "add first node" ran with condition "service_completed_successfully"
    And the scenario "add second node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | functionPort | firstHandle  | secondHandle         | secondPort |
      | my-solution | my-workflow    | first-function | test-port    | first-handle | second-function-node | input      |
    And the workdir changes to "<folder>"
    When I run the command "wf links add <workflowHandle> <functionFolder>/run/out/<functionPort>:<secondHandle>[input]"
    Then I should have a link under "<folder>", workflow "<workflowHandle>" with left side handle "<firstHandle>", port "<functionPort>" and a right side handle "<secondHandle>" on port

  Scenario: remove link
    Given the scenario "link first and second nodes" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | functionPort | firstHandle  | secondHandle         | secondPort |
      | my-solution | my-workflow    | first-function | test-port    | first-handle | second-function-node | input      |
    And the workdir changes to "<folder>"
    When I run the command "wf links rm <workflowHandle> <functionFolder>/run/out/<functionPort>:<secondHandle>[input]"
    Then I should have no links under "<folder>", workflow "<workflowHandle>"