Feature: function test with property of bash example
  To be able to test the bash example function with the property MY_KEY
  As a local user
  I should be able to create and build the bash example
  I should be able to create a property specification for MY_KEY
  I should be able to add an environment variable for MY_KEY
  I should be able to test the function

  Scenario: create bash example
    Given the following parameters
      | path               | oem            | version |
      | prop-bash-function | com.genaiz.dev | 1.0.0   |
    When I run the command "sf create <path> --oem=<oem> --recipe=bash-example"
    Then I should have a function under "<path>" named "<path>", handle "<path>", oem "<oem>" with version "<version>"
    And I should have dockerfile under "<path>" named "Dockerfile"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | repository                        | version |
      | prop-bash-function | com.genaiz.dev/prop-bash-function | latest  |
    And the workdir changes to "<path>"
    When I run the command "sf build"
    Then I should have a docker image under repository "<repository>" with tag "<version>"

  Scenario: add prop key
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | key    | name        | defaultValue | type |
      | prop-bash-function | MY_KEY | Key Example | 10           | int  |
    And the workdir changes to "<path>"
    When I run the command "sf prop add <key> --name='<name>' --default-value=<defaultValue> --type=<type>"
    Then I should have a property specification for key "<key>" under path "<path>"

  Scenario: define key env var
    Given the scenario "add prop key" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | key    | value |
      | prop-bash-function | MY_KEY | 12    |
    And the workdir changes to "<path>"
    When I run the command "sf prop env <key> 12"
    Then I should have a property specification for key "<key>" under path "<path>"

  Scenario: test bash example
    Given the scenario "define key env var" ran with condition "service_completed_successfully"
    And the following parameters
      | path               | oem            | handle             | version | key    | value |
      | prop-bash-function | com.genaiz.dev | prop-bash-function | latest  | MY_KEY | 12    |
    And the workdir changes to "<path>"
    When I run the command "sf test"
    Then I should have a test session attached with the test output
    And I should have an output for key "<key>" with value "<value>"
