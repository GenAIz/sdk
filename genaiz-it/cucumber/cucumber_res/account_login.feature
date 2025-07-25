Feature: account login
  To be able to authenticate
  As a regular user
  I need to be able to create a session

  Scenario: create session ok
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=success"
    And the following parameters
      | username         |
      | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have a session id with host "<orchestrator>" for username "<username>"
