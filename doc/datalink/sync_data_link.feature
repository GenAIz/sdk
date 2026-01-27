Feature: data link sync
  To be able to sync a data link
  As an authenticated developer
  I should be able to create the data link locally, publish it to an orchestrator and export it to another local user path

  Scenario: create data link locally
    Given the following parameters
      | path       | configFile             | handle     | oem            | version |
      | my-project | my-project/Genaiz.yaml | datalink-1 | com.genaiz.dev | 0.2.1   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> <path> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link double property
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | path       | configFile             | handle     | oem            | version | key      | type   | defaultValue |
      | my-project | my-project/Genaiz.yaml | datalink-1 | com.genaiz.dev | 0.2.1   | TEST_KEY | double | 13.0         |
    And the workdir changes to "<path>"
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --type=<type> --default-value=<defaultValue> --user-defined=false"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

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
    And the following parameters
      | path       | handle     | oem            | version |
      | my-project | datalink-1 | com.genaiz.dev | 0.2.1   |
    When I run the command "dk publish <oem>/<handle> <path> --version=<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: synchronize data link to user configuration
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 0.2.1   | TEST_KEY |
    When I run the command "dk sync <oem>/<handle> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"
    And I should have property under "<configFile>", for data link "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: edit data link double property
    Given the scenario "synchronize data link to user configuration" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      | defaultValue | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 0.2.1   | TEST_KEY | 42.2         | DOUBLE |
    When I run the command "dk prop edit <oem>/<handle>:<version <key> --default-value=<defaultValue>
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

  Scenario: publish data link new revision
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem            | version | newVersion |
      | datalink-1 | com.genaiz.dev | 0.2.1   | 0.2.2            |
    When I run the command "dk publish <oem>/<handle>:<version> --new-version=<newVersion>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<newVersion>"

  Scenario: synchronize data link new revision to user configuration
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      | oldVersion |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 0.2.2   | TEST_KEY | 0.2.1      |
    When I run the command "dk sync <oem>/<handle> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"
    And I should have property under "<configFile>", for data link "<handle>", oem "<oem>" and version "<version>", with key "<key>"
    And I should not have a datalink under "<configFile>" named "<handle>" with handle "<handle>", oem "<oem>" and version "<oldVersion>"
