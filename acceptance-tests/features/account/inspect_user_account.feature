Feature: account login with username
  To be able to authenticate
  As a regular user
  I need to be able to create a session

  Scenario: create session
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: inspect session
    Given the scenario "create session" ran with condition "service_completed_successfully"
    When I run the command "ac inspect --json"
    Then I should have a unexpired session with a session id

  Scenario: inspect session from environment
    Given the scenario "create session" ran with condition "service_completed_successfully"
    And the environment contains "GENAIZ_AUTH_URL=<orchestrator>"
    And the environment contains "GENAIZ_AUTH_SESSION=<sessionToken>"
    And the parameter "sessionToken" is taken from the auth file "<path>" for broker "<orchestrator>"
    And the following parameters
      | path                      |
      | $HOME/.cache/genaiz/.auth |
    When I run the command "ac inspect --json"
    Then I should have a unexpired session with a session id
