Feature: account login with OIDC
  To be able to authenticate
  As an OIDC user
  I need to be able to create a device url, authorize it through an IDP and create a session with the IDP token

  Scenario: create oidc session no browser ok
    Given the orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestrator> --no-browser"
    And I open the oidc url provided on the shell onto a browser
    And I click "Sign In" with the username "<username>" into the email field
    And I click "Sign in" with the password "<password>" into the password field
    And I answer the two factor authentification challenge if any is presented
    And I click on "Yes" for the given access privileges
    And I hit enter on the shell
    Then I should have an active session id with host "<orchestrator>" for username "token" under path "<path>"
    And I should see "Logged in to <orchestrator>" on the shell

  Scenario: create oidc session pending authorization
    Given the orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestrator>"
    And I hit enter on the shell
    Then I should see the oidc url provided twice

  Scenario: create oidc session throttling authorization
    Given the orchestrator is running with condition: "service_healthy"
    And the oidc provider is running with condition: "service_healthy"
    And the following parameters
      | path                      | username         | password     |
      | $HOME/.cache/genaiz/.auth | _test@genaiz.com | testPassword |
    When I run the command "ac login <orchestrator>"
    And I hit enter on the shell
    And I hit enter on the shell
    Then I should see an error "slow_down"

  Scenario create oidc session unsupported
    Given the orchestrator is running with condition: "service_healthy"
    And the oidc provider urls are not configured
    When I run the command "ac login <orchestrator>"
    Then I should see a prompt for a username
