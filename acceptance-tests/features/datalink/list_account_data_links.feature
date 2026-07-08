Feature: list account data links
  To be able to list data links
  As an authenticated developer
  I should be able to create, publish and list a data link

  Scenario: create data link
    Given the following parameters
      | configFile                       | handle     | oem             | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: login data link
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish data link
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem             | version |
      | datalink-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: list data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the following parameters
      | oem             | handle     | version | seq |
      | com.genaiz.test | datalink-1 | 1.0.0   | 0   |
    When I run the command "dk list <fqdnPrefix> --account-only --json"
    Then I should have a list containing the data link "<oem>/<handle>:<version>" with sequence "<seq>"
