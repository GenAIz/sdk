Feature: function nodes for an extensive workflow
  To be able to add and remove nodes from a workflow
  As a developer
  I should be able to create a solution, create a function, add workflow nodes and remove some

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem             | handle     | version | workflowHandle | workflowName   | workflowDescription  |
      | my-solution | com.genaiz.test | solution-1 | 0.1.1   | workflow-1     | First Workflow | Workflow Description |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --version=<version --workflow-handle=<workflowHandle> --workflow-name='<workflowName>' --workflow-desc='<workflowDescription>'"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create bash example
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem             | version | type     |
      | my-solution | bash-example | my-function | com.genaiz.test | 1.0.0   | function |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add node to invalid workflow
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionHandle | workflowHandle   | oem             | version |
      | my-solution | my-function    | invalid-workflow | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionHandle>/
    Then I should have an error with "workflow hande [<workflowHandle>] not found"

  Scenario: add serialized node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionHandle | workflowHandle | nodeHandle | oem             | version |
      | my-solution | my-function    | workflow-1     | my-node    | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf=<oem>/<functionHandle>:<version>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add external node
    Given the scenario "add serialized node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle       | sfOem           | sfHandle    | sfVersion | sfSeq |
      | my-solution | workflow-1     | my-external-node | com.genaiz.test | my-external | 1.0.0     | 2     |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf-oem=<sfOem> --sf-handle=<sfHandle> --sf-version=<sfVersion> --sf-seq=<sfSeq>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<sfOem>", handle "<sfHandle>", version "<sfVersion>" and sequence "<sfSeq>"

  Scenario add duplicate serialized node
    Given the scenario "add serialized node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionFolder | workflowHandle | nodeHandle | oem             | version |
      | my-solution | my-function    | workflow-1     | my-node    | com.genaiz.test | 1.0.0   |
    When I run the command "wf nodes add <workflowHandle> <nodeHandle>"
    Then I should have an error with "the node specified already exists"

  Scenario: remove node from invalid workflow
    Given the scenario "add duplicate serialized node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle | validWorkflow |
      | my-solution | invalid-handle | my-node    | workflow-1    |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes rm <workflowHandle> <nodeHandle>"
    Then I should have an error with "workflow hande [<workflowHandle>] not found"
    And I should have a node under "<folder>" and workflow "<validWorkflow> named "<nodeHandle>" and handle "<nodeHandle>"

  Scenario: remove invalid node
    Given the scenario "remove node from invalid workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle   | validHandle |
      | my-solution | workflow-1     | invalid-node | my-node     |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes rm <workflowHandle> <nodeHandle>"
    Then I should not have an error
    And I should have a node under "<folder>" and workflow "<workflowHandle> named "<validHandle>" and handle "<validHandle>"

  Scenario: remove external node
    Given the scenario "add external node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle | externalHandle |
      | my-solution | workflow-1     | my-node    | external-node  |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes rm <workflowHandle> <nodeHandle>"
    Then I should not have a node under "<folder>" and workflow "<workflowHandle>" with handle "<nodeHandle>"
    And I should have a node under "<folder>" and workflow "<workflowHandle> named "<externalHandle>" and handle "<externalHandle>"
