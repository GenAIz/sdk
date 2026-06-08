Feature: account activation with OIDC
  To be able to activate an OIDC session
  As an OIDC user
  I need to be able to login onto an OIDC broker, login onto a second OIDC broker, and switch session for the first one

  Scenario: create first oidc session
    Given the orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestrator>"
    And I click "Sign In" with the username "<username>" into the email field
    And I click "Sign in" with the password "<password>" into the password field
    And I answer the two factor authentification challenge if any is presented
    And I click on "Yes" for the given access privileges
    Then I should have an active session id with host "<orchestrator>" for username "token" under path "<path>"
    And I should see "Logged in to <orchestrator>" on the shell

  Scenario: create second oidc session
    Given the scenario "create first oidc session" ran with condition "service_completed_successfully"
    And the alternate orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestrator_2>"
    And I click "Sign In" with the username "<username>" into the email field
    And I click "Sign in" with the password "<password>" into the password field
    And I answer the two factor authentification challenge if any is presented
    And I click on "Yes" for the given access privileges
    Then I should have an active session id with host "<orchestrator_2>" for username "token" under path "<path>"
    And I should see "Logged in to <orchestrator_2>" on the shell

  Scenario: list oidc accounts
    Given the scenario "create second oidc session" ran with condition "service_completed_successfully"
    When I run the command "ac list"
    Then I should have an active session with host "<orchestrator_2>" with a username, a created date and an expiry date

  Scenario: activate first oidc account
    Given the scenario "list oidc accounts" ran with condition "service_completed_successfully"
    When I run the command "ac activate <orchestrator>"
    Then I should see "Activating session with <orchestrator>"

  Scenario: list activated oidc accounts
    Given the scenario "activate first oidc account" ran with condition "service_completed_successfully"
    When I run the command "ac list"
    Then I should have an active session with host "<orchestrator>" with a username, a created date and an expiry date
