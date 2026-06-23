Feature: function test with properties of a bash connector datalink
  To be able to test the bash example function with a datalink property
  As an authenticated user
  I should be able to create and build the bash example
  I should be able to create a datalink with a property spec for DL_KEY
  I should be able to add an environment variable for DL_KEY
  I should be able to create a datalink with a secret property spec for SECRET_KEY
  I should be able to add an environment variable for SECRET_KEY

  Scenario: create bash example
    Given the following parameters
      | path        | oem            | version | type      |
      | my-function | com.genaiz.test | 1.0.0   | connector |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example --type=<type>"
    Then I should have a function under "<handle>" named "<name>" with oem "<oem>", version "<version>" and type "<type>"

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
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link property
    Given the scenario "create data link for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key    | type | value |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   | DL_KEY | INT  | 37    |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --type=<type> --default-value=<value>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<value>"

  Scenario: add data link secret property
    Given the scenario "add data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key        | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   | SECRET_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: publish data link for bash example
    Given the scenario "add data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem            | version |
      | datalink-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator

  Scenario: add data source to function
    Given the scenario "publish data link for bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | handle     | oem            | version |
      | my-function | datalink-1 | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: define datalink key invalid env var
    Given the scenario "add data source to function" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | key    | value        | type |
      | my-function | DL_KEY | string_value | INT  |
    And the workdir changes to "<path>"
    When I run the command "sf prop env <key> <value>"
    Then I should have an error with "illegal <type> value"

  Scenario: define datalink key env var
    Given the scenario "add data source to function" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | key    | value |
      | my-function | DL_KEY | 42    |
    And the workdir changes to "<path>"
    When I run the command "sf prop env <key> <value>"
    Then I should have a .env file under "<path>" with key "<key>" and value "<value>"

  Scenario: define datalink secret key env var
    Given the scenario "define datalink key env var" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | key        | value  |
      | my-function | SECRET_KEY | secret |
    And the workdir changes to "<path>"
    When I run the command "sf prop env <key> <value>"
    Then I should have a .env file under "<path>" with key "<key>" and value "<value>"
