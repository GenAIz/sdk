Feature: list accounts for OIDC and regular usernames
  To be able to list accounts with sessions
  As a regular user
  I need to be able to create a session with a username, and a session with an IDP token, then list them

  Scenario: create oidc session
    Given the OIDC orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestratorOidc>"
    And I open the oidc url provided on the shell onto a browser
    And I click "Sign In" with the username "<username>" into the email field
    And I click "Sign in" with the password "<password>" into the password field
    And I answer the two factor authentification challenge if any is presented
    And I click on "Yes" for the given access privileges
    Then I should have an active session id with host "<orchestratorOidc>" for username "token" under path "<path>"
    And I should see "Logged in to <orchestratorOidc>" on the shell

  Scenario: create session
    Given the orchestrator is running with condition: "service_healthy"
    And the scenario "create oidc session" ran with condition "service_completed_successfully"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: list sessions json
    Given the scenario "create session" ran with condition "service_completed_successfully"
    And the following parameters
      | username         |
      | _test@genaiz.com |
    When I run the command "ac list --json"
    Then I should have an account session for user "<username>" on host "<orchestrator>"
    And I should have an account session on host "<orchestratorOidc>"
    And I should have an active account session for host "<orchestrator>"

  Scenario: login to existing session
    Given the scenario "list session json" ran with condition "service_completed_successfully"
    When I run the command "ac login <orchestratorOidc>"
    Then I should see "Already logged in to <orchestratorOidc>"

  Scenario: list active session json
    Given the scenario "login to existing session" ran with condition "service_completed_successfully"
    When I run the command "ac list --json"
    Then I should have an active account session for host "<orchestratorOidc>"
