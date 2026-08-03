Feature: workflow create for a simple solution
  To be able to create a simple workflow
  As a developer
  I should be able to create a solution, create a new workflow, create a function and add a workflow node

  Scenario: create simple solution with defaults
    Given the following parameters
      | folder      | oem             | handle     | version | workflowHandle | workflowName     | workflowDescription |
      | my-solution | com.genaiz.test | solution-1 | 1.0.0   | default        | Default Workflow | default workflow    |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle>"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create simple workflow
    Given the scenario "create simple solution with defaults" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      |
      | my-solution | my-workflow |
    When I run the command "wf create <handle> <folder>"
    Then I should have workflow under "<folder>" named "<handle>", handle "<handle>" with description "<handle>"

  Scenario: create bash example
    Given the scenario "create simple solution with defaults" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem             | version | type     |
      | my-solution | bash-example | my-function | com.genaiz.test | 1.0.0   | function |
    When I run the command "sf create <handle> --context=<folder> --recipe=<recipe>"
    Then I should have a function under "<folder>/<handle>" named "<handle>" with oem "<oem>" and version "<version>" of type "<type>"

  Scenario: add bash example node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | nodeHandle       | oem             | version |
      | my-solution | my-workflow    | my-function    | my-function-node | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<functionFolder>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add data port to bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                  | portFolder | portHandle | portName  | portDesc      | workflowHandle |
      | my-solution/my-function | run/out    | test-port  | Test Port | A description | my-workflow    |
    And the workdir changes to "<folder>"
    When I run the command "sf data output add <portFolder>/<portHandle> --name='<portName>' --description='<portDesc>'"
    Then I should have an output port under "<folder>" named "<portName>" with handle "<portHandle>" and description "<portDesc>"

  Scenario: add external node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | nodeHandle       | sfOem           | sfHandle    | sfVersion | sfSeq |
      | my-solution | my-workflow    | my-external-node | com.genaiz.test | my-external | 1.0.0     | 2     |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf-oem=<sfOem> --sf-handle=<sfHandle> --sf-version=<sfVersion> --sf-seq=<sfSeq>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<sfOem>", handle "<sfHandle>", version "<sfVersion>" and sequence "<sfSeq>"

  Scenario: add bash example output link
    Given the scenario "add external node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle | functionFolder | functionPort | functionNodeHandle | externalHandle   | externalPort |
      | my-solution | my-workflow    | my-function    | test-port    | my-function-node   | my-external-node | test-1       |
    And the workdir changes to "<folder>"
    When I run the command "wf links add <workflowHandle> <functionFolder>/run/out/<functionPort>:<externalHandle>[<externalPort>]
    Then I should have a link under "<folder>", workflow "<workflowHandle>" with left side handle "<functionNodeHandle>", port "<functionPort>" and a right side handle "<externalHandle>"
