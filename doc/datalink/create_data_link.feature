Feature: data link create
  To be able to create a data link
  As an authenticated developer
  I should be able to create and publish the data link

  Scenario: create bash example
    Given the scenario "create bash example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | recipe       | handle            | oem            | name              | type      | version |
      | bash-example | my-bash-connector | com.genaiz.dev | My Bash Connector | connector | 0.1.1   |
    When I run the command "sf create <handle> --oem=<oem> --name='<name>' --recipe=<recipe> --type=<type>"
    Then I should have a function under "<handle>" named "<name>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: create data link invalid handle
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     |
      | my-bash-example | not--valid |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle>"
    Then I should have an error for field "datalink.create.handle"

  Scenario: create data link invalid oem
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem        |
      | my-bash-example | datalink-1 | not.valid. |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have an error for field "datalink.create.oem"

  Scenario: create data link invalid version
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-1 | com.genaiz.dev | 1..0    |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "datalink.create.version"

  Scenario: login bash example
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: create data link for bash example
    Given the scenario "login bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: publish data link for bash example
    Given the scenario "create data link for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem            | version |
      | datalink-1 | com.genaiz.dev | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator

  Scenario: add data source to bash example
    Given the scenario "create data link for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"
