Feature: solution publish with validation
  To validate a solution before publishing
  As an authenticated user
  I should be able to create a solution, create a function, modify its metadata and fail publishing it with validation errors

  Scenario: create solution to validate
    Given the following parameters
      | folder      | oem            | version | workflowDesc     | workflowHandle | workflowName     |
      | my-solution | com.genaiz.dev | 1.0.0   | default workflow | default        | Default Workflow |
    When I run the command "sn create <folder> --oem=<oem>"
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create node function
    Given the following parameters
      | folder      | recipe       | handle      | oem            | type      | version |
      | my-solution | bash-example | my-function | com.genaiz.dev | connector | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe> --type=<type>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: login solution validation
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: create data link for solution validation
    Given the scenario "login function validation" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: publish data link for solution validation
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem            | version |
      | datalink-1 | com.genaiz.dev | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: add data source to bash example function
    Given the scenario "publish data link for function validation" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle     | oem            | version |
      | my-function | datalink-1 | com.genaiz.dev | 1.0.0   |
    When I run the command "sf data src add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  # There's definitely a point to make that this shouldn't be possible to begin, but since one can always edit files
  # with a text editor, the supplementary efforts to frame init behind a wall of validation would just not pay
  Scenario re-initialize the function as type function
    Given the scenario "add data source to bash example function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | type     | srcDataLink                      |
      | my-function | function | com.genaiz.dev/data-source:1.0.0 |
    When I run the command "sf init --type=<type>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"
    And I should have a data source under "<folder>" with datalink "<srcDataLink>"

  Scenario: add workflow node to simple solution
    Given the scenario "build bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | function    | workflowHandle | handle | description | functionOem    | functionVersion |
      | my-solution | my-function | default        | node-1 |             | com.genaiz.dev | 1.0.0           |
    And the workdir changes to "<solution>"
    When I run the command "wf nodes add <workflowHandle> <handle>/"
    Then I should have a workflow node under "<solution>" with handle "<handle>", oem "<oem>", description "<description>" and smart function "<functionOem>/<function>:<functionVersion>"

  Scenario publish solution type failure
    Given the scenario "re-initialize the function as type function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | type     |
      | my-function | function |
    And the workdir changes to "<folder>"
    When I run the command "sn publish"
    Then I should I should get the error "Error: type [<type>] can not specify data sources"
