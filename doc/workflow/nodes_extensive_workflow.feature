Feature: function nodes for an extensive workflow
  To be able to add and remove nodes from a workflow
  As a developer
  I should be able to create a solution, create a function, add workflow nodes and remove some

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem            | handle     | version | workflowHandle | workflowName   | workflowDescription  |
      | my-solution | com.genaiz.dev | solution-1 | 0.1.1   | workflow-1     | First Workflow | Workflow Description |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --version=<version --workflow-handle=<workflowHandle> --workflow-name='<workflowName>' --workflow-desc='<workflowDescription>'"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create bash example
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem            | version |
      | my-solution | bash-example | my-function | com.genaiz.dev | 0.1.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add serialized node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionHandle | workflowHandle | nodeHandle | oem            | version |
      | my-solution | my-function    | my-workflow    | my-node    | com.genaiz.dev | 0.1.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf=<oem>/<functionHandle>:<version>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<functionFolder>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add external node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle       | sfOem          | sfHandle    | sfVersion | sfSeq |
      | my-solution | my-workflow    | my-external-node | com.genaiz.dev | my-external | 0.1.0     | 2     |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf-oem=<sfOem> --sf-handle=<sfHandle> --sf-version=<sfVersion> --sf-seq=<sfSeq>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<sfOem>", handle "<sfHandle>", version "<sfVersion>" and sequence "<sfSeq>"

  Scenario: remove serialized node
    Given the scenario "add external node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle |
      | my-solution | my-workflow    | my-node    |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes rm <workflowHandle> <functionFolder>"
    Then I should not have a node under "<folder>" and workflow "<workflowHandle>" with handle "<nodeHandle>"