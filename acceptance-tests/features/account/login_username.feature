Feature: account login with username
  To be able to authenticate
  As a regular user
  I need to be able to create a session

  Scenario: create session ok
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario create session existing ok
    Given the scenario "create session ok" ran with condition "service_completed_successfully"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the session id under <path> for username <username> and orchestrator <orchestrator>
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have the same active session id "<sessionId>" for username "<username>" and orchestrator "<orchestrator>" under path "<path>"

  Scenario create session refresh ok
    Given the scenario "create session existing ok" ran with condition "service_completed_successfully"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the session id under <path> for username <username> and orchestrator <orchestrator>
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username> --refresh"
    Then I should have a different active session id "<sessionId>" for username "<username>" and orchestrator "<orchestrator>" under path "<path>"

  Scenario create session multi ok
    Given the scenario "create session refresh ok" ran with condition "service_completed_successfully"
    And the second orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the session id under <path> for username <username> and orchestrator <orchestrator>
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator_2> --username=<username>"
    Then I should have an active session id with host "<orchestrator_2>" for username "<username>" under path "<path>"
    And I should have an inactive session id with host "<orchestrator>" for username "<username>" under path "<path"

  Scenario activate first session
    Given the scenario "create session multi ok" ran with condition "service_completed_successfully"
    And the following parameters
      | path                      | username         |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario activate second session
    Given the scenario "create session multi ok" ran with condition "service_completed_successfully"
    And the following parameters
      | path                      | username         |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator_2>"
    Then I should have an active session id with host "<orchestrator_2>" for username "<username>" under path "<path>"

  Scenario delete session ok
    Given the scenario "create session multi ok" ran with condition "service_completed_successfully"
    And the second orchestrator is running with condition: "service_healthy"
    And the session id under <path> for username <username> and orchestrator <orchestrator_2>
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac logout"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"
    And I should not have an active session id with host "<orchestrator_2>" for username "<username>"

  Scenario delete session with username ok
    Given the scenario "delete session ok" ran with condition "service_completed_successfully"
    And the second orchestrator is running with condition: "service_healthy"
    And the session id under <path> for username <username> and orchestrator <orchestrator>
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac logout --username <username>"
    Then I should not have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"
