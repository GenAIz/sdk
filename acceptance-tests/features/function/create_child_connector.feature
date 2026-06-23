Feature: connector create for a parent solution
  To be able to create a connector function as a child of a parent solution
  As a developer
  I should be able to create the alpine default recipe and edit its metadata

  Scenario: create parent solution
    Given the following parameters
      | folder          | oem            | handle     | description          | name        | version | workflowDesc     | workflowHandle | workflowName     |
      | parent-solution | com.genaiz.test | solution-1 | solution description | My Solution | 0.1.1   | default workflow | default        | Default Workflow |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --description='<description>' --name='<name>' --version=<version>""
    Then I should have a solution under "<folder>" named "<name>" with oem "<oem>", handle "<handle>", description "<description>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create child function
    Given the scenario "create parent solution" ran with condition "service_completed_successfully"
    And the following parameters
      | solutionFolder  | folder         | oem            | version | type     | dockerFile | dockerIgnore  |
      | parent-solution | child-function | com.genaiz.test | 0.1.1   | function | Dockerfile | .dockerignore |
    When I run the command "sf create <folder> --solution-path=<solutionFolder>"
    Then I should have a function under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", version "<version>" and type "<type>"
    And I should have a file "<dockerFile>" under "<folder>"
    And I should have a file "<dockerIgnore>" under "<folder>"

  Scenario change child function type to connector
    Given the scenario "create child function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder         | type      | oem            | version |
      | child-function | connector | com.genaiz.test | 0.1.1   |
    And the workdir changes to "<folder>"
    When I run the command "sf init --type=connector"
    Then I should have a function under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", version "<version>" and type "<type>"

  Scenario: create data link for child function
    Given the following parameters
      | configFile                       | handle     | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.1   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link secret property
    Given the scenario "create data link for child function" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key        | type |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.1   | SECRET_KEY | int  |
    When I run the command "dk prop add <oem>/<handle> <key> --version=<version> --type=<type> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: login data link
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish data link
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the scenario "add data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem            | version |
      | datalink-1 | com.genaiz.test | 1.0.1   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: add data source to child function
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the following parameters
      | folder         | handle     | oem            | version | connectorVersion | type      |
      | child-function | datalink-1 | com.genaiz.test | 1.0.0   | 0.1.1            | connector |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"
    And I should have a function under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", version "<connectorVersion>" and type "<type>"

  Scenario: add property spec to child function
    Given the scenario "add data source to child function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder         | key    | defaultValue | type   |
      | child-function | MY_KEY | value        | string |
    And the workdir changes to "<path>"
    When I run the command "sf prop add <key> --default-value=<defaultValue>"
    Then I should have a property specification for key "<key>", of type "<type>" under path "<folder>"
    And I should have a function under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", version "<connectorVersion>" and type "<type>"
