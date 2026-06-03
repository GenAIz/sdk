Feature: create simple workspace
  To be able to create a simple workspace
  As a developer
  I should be able to login to a broker, and then create a private or organization workspace

  Scenario: create workspace invalid account
    Given the following parameters
      | name           | account         |
      | test_workspace | invalid.account |
    When I run the command "ws create <name> --account=<account>"
    Then I should have an error "Error: could not elect a session"

  Scenario: create session ok
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: create workspace with support for release candidates
    Given the scenario "create session ok" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 3     |
    When I run the command "ws create <name> --json"
    Then I should have a workspace with name "<name>", a created timestamp the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace for organization
    Given the scenario "create session ok" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | description      | visibility   | flags |
      | test_workspace | test description | ORGANIZATION | 1     |
    When I run the command "ws create <name> --description='<test_desc>' --visibility=<visibility> --rc-enabled=false --json"
    Then I should have a workspace with name "<name>", a created timestamp the visibility set to "<visibility>" and flags set to "<flags>"
