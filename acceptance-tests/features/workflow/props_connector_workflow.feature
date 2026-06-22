Feature: workflow node properties for a connector workflow
  To be able to manage properties on a workflow connector node
  As a developer
  I should be able to create a solution, create a function with an associated node, create local data links,
  add the connector to the function, and add, remove and edit properties on this node

  Scenario: create basic solution
    Given the following parameters
      | folder      | oem            | version | workflowHandle | workflowName     | workflowDescription |
      | my-solution | com.genaiz.test | 1.0.0   | default        | Default Workflow | default workflow    |
    When I run the command "sn create <folder> --oem=com.genaiz.test"
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create bash example
    Given the scenario "create basic solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | recipe       | handle      | oem            | version | type      |
      | my-solution | bash-example | my-function | com.genaiz.test | 1.0.1   | connector |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --version=<version> --type=<type> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add bash example node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | functionFolder | workflowHandle | nodeHandle       | oem            | version |
      | my-solution | my-function    | default        | my-function-node | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <functionFolder>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: create data source data link
    Given the following parameters
      | configFile                       | handle       | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-src | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data source data link property
    Given the scenario "create data source data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem            | version | key     | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-src | com.genaiz.test | 1.0.0   | SRC_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value ""

  Scenario: create data store data link
    Given the scenario "add data source data link property" ran with condition "service_completed_successfully"
    Given the following parameters
      | configFile                       | handle       | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-str | com.genaiz.test | 1.0.1   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data store data link property
    Given the scenario "create data store data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem            | version | key     | type | defaultValue |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-str | com.genaiz.test | 1.0.1   | STR_KEY | INT  | 37           |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --type=<type> --default-value=37"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

  Scenario: add data source to bash example
    Given the scenario "add data source data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                  | handle       | oem            | version |
      | my-solution/my-function | datalink-src | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version> --no-validation"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add data store to bash example
    Given the scenario "add data source to bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                  | handle       | oem            | version |
      | my-solution/my-function | datalink-str | com.genaiz.test | 1.0.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version> --no-validation"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add data source prop to bash example node
    Given the scenario "add data source to bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle  | nodeHandle       | key     | value       |
      | my-solution | default | my-function-node | SRC_KEY | value_added |
    And the workdir changes to "<folder>"
    When I run the command "wf prop add <handle> <nodeHandle> <key> '<value>' --no-prop-sync"
    Then I should have a property on workflow node "<nodeHandle>" under workflow "<handle>" with key "<key>" and value "<value>"
