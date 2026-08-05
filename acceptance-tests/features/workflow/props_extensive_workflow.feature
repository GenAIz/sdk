Feature: workflow node properties for an extensive workflow
  To be able to manage properties on a workflow node
  As a developer
  I should be able to create a solution, create a new workflow, create a function with an associated node, and add, remove and edit properties on this node

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem             | handle     | version | workflowHandle | workflowName     | workflowDescription |
      | my-solution | com.genaiz.test | solution-1 | 1.0.0   | default        | Default Workflow | default workflow    |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle>"
    Then I should have a solution under "<folder>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: add prop to invalid workflow
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle           | nodeHandle   | key | value |
      | my-solution | invalid-workflow | invalid-node | KEY | value |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>'"
    Then I should I should get the error "Error: workflow hande [<handle>] not found"

  Scenario: create additional workflow
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | workflowName | workflowDescription  |
      | my-solution | my-workflow | My Workflow  | Workflow Description |
    And the workdir changes to "<folder>"
    When I run the command "wf create <handle> --name='<workflowName>' --description='<workflowDescription>'"
    Then I should have workflow under "<folder>" named "<workflowName>", handle "<handle>" with description "<workflowDescription>"

  Scenario: create bash example
    Given the scenario "create additional workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem             | version | type     |
      | my-solution | bash-example | my-function | com.genaiz.test | 1.0.0   | function |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --oem=<oem> --version=<version> --type=<type> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add bash example prop spec
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path                    | key    | name        | defaultValue | type |
      | my-solution/my-function | MY_KEY | Key Example | False        | bool |
    And the workdir changes to "<path>"
    When I run the command "sf prop add <key> --default-value=<defaultValue> --type=<type>"
    Then I should have a property specification for key "<key>" under path "<path>" with type "<type>", and default value "<defaultValue>"

  Scenario: add bash example node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionHandle | workflowHandle | nodeHandle | oem             | version |
      | my-solution | my-function    | my-workflow    | my-node    | com.genaiz.test | 1.0.0   |
      | my-solution | my-function    | default        | my-node    | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf=<oem>/<functionHandle>:<version>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<functionFolder>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: add prop to invalid node
    Given the scenario "create additional workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle   | key    | value |
      | my-solution | my-workflow | invalid-node | MY_KEY | value |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>'"
    Then I should get the error "Error: node [<nodeHandle>] is not a member of workflow [<handle>]"

  Scenario: add invalid prop key to node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key         | value |
      | my-solution | my-workflow | my-node    | INVALID_KEY | value |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>'"
    Then I should get the error "Error: node [<nodeHandle>] is not a member of workflow [<handle>]"

  Scenario: add prop to bash example node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key    | value       |
      | my-solution | my-workflow | my-node    | MY_KEY | value_added |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>'"
    Then I should have a property on workflow node "<nodeHandle>" under workflow "<handle>" with key "<key>" and value "<value>"

  Scenario: edit invalid prop from bash example node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key     | value |
      | my-solution | my-workflow | my-node    | INVALID | value |
    And the workdir changes to "<folder>"
    When I run the command "wf prop edit <handle> <nodeHandle> <key> '<value>'"
    Then I should get the error "Error: the key [<key>] could not be found under node [<nodeHandle>]"

  Scenario: add duplicate prop to bash example node
    Given the scenario "add prop to bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key    | value   |
      | my-solution | my-workflow | my-node    | MY_KEY | invalid |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>'"
    Then I should get the error "Error: the key [<key>] is already defined for node [<nodeHandle>]"

  Scenario: edit prop from bash example node
    Given the scenario "add bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key    | value        |
      | my-solution | my-workflow | my-node    | MY_KEY | value_edited |
    And the workdir changes to "<folder>"
    When I run the command "wf prop edit <handle> <nodeHandle> <key> '<value>'"
    Then I should have a property on workflow node "<nodeHandle>" under workflow "<handle>" with key "<key>" and value "<value>"

  Scenario: add bash example unused prop spec
    Given the scenario "add bash example prop spec" ran with condition "service_completed_successfully"
    And the following parameters
      | path                    | key       | name       | type   |
      | my-solution/my-function | OTHER_KEY | Unused Key | STRING |
    And the workdir changes to "<path>"
    When I run the command "sf prop add <key>"
    Then I should have a property specification for key "<key>" under path "<path>" with type "<type>"

  Scenario: edit unused prop from bash example node
    Given the scenario "add bash example unused prop spec" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key       | value         |
      | my-solution | my-workflow | my-node    | OTHER_KEY | value_invalid |
    And the workdir changes to "<folder>"
    When I run the command "wf prop edit <handle> <nodeHandle> <key> '<value>'"
    Then I should get the error "Error: the key [<key>] could not be found under node [<nodeHandle>]"

  Scenario remove prop from bash example node
    Given the scenario "add prop to bash example node" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle      | nodeHandle | key    |
      | my-solution | default     | my-node    | MY_KEY |
      | my-solution | my-workflow | my-node    | MY_KEY |
    And the workdir changes to "<folder>"
    When I run the command "wf prop rm <handle> <nodeHandle> <key>"
    Then I should not have a property on workflow node "<nodeHandle>" under workflow "<handle>" with key "<key>"
