Feature: list user workspaces
  To list an account workspaces
  As a developer
  I should be able to login to a broker, create a workspace and list it

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
    Then I should have a workspace with name "<name>", a created timestamp, the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: list workspaces with support for release candidates
    Given the scenario "create workspace with support for release candidates" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 3     |
    When I run the command "ws list --json"
    Then I should have a list of workspaces with a workspace named "<name>", a created timestamp, the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace without support for release candidates
    Given the scenario "create session ok" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 1     |
    When I run the command "ws create <name> --rc-enabled=false --json"
    Then I should have a workspace with name "<name>", a created timestamp the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: list workspaces with support for release candidates
    Given the scenario "create workspace with support for release candidates" ran with condition "service_completed_successfully"
    And the following parameters
      | name           | visibility | flags |
      | test_workspace | PRIVATE    | 1     |
    When I run the command "ws list --rc-enabled=false --json"
    Then I should have a list of workspaces with a workspace named "<name>", a created timestamp, the visibility set to "<visibility>" and flags set to "<flags>"
