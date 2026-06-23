Feature: list simple solutions
  To be able to list solutions
  As a dev
  I need to be able to create a solution, and then list it

  Scenario: create simple solution with name and description
    Given the following parameters
      | folder      | oem            | version | solutionName | solutionDesc         | workflowName     | workflowHandle | workflowDesc     |
      | my-solution | com.genaiz.test | 1.0.0   | My Solution  | Solution Description | Default Workflow | default        | default workflow |
    When I run the command "sn create <folder> --oem=<oem> --name='<solutionName>' --description='<solutionDesc>'"
    Then I should have a solution under "<folder>" named "<solutionName>" with oem "<oem>", handle "<folder>", description "<solutionDesc>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: list simple solution no argument
    Given the scenario "create simple solution with name and description" ran with condition "service_completed_successfully"
    And the following parameters
      | oem            | handle      | version | solutionName | local |
      | com.genaiz.test | my-solution | 1.0.0   | My Solution  | true  |
    When I run the command "sn list --json"
    Then I should have a solution list with solution "<oem>/<handle>:<version>" named "<solutionName>" and local flag set to "<local>"

  Scenario: list simple solution with folder path
    Given the scenario "create simple solution with name and description" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | oem            | handle      | version | solutionName | local |
      | my-solution | com.genaiz.test | my-solution | 1.0.0   | My Solution  | true  |
      | .           | com.genaiz.test | my-solution | 1.0.0   | My Solution  | true  |
    When I run the command "sn list <folder> --json"
    Then I should have a solution list with solution "<oem>/<handle>:<version>" named "<solutionName>" and local flag set to "<local>"

  Scenario: create sub-folder solution
    Given the following parameters
      | folder    | oem            | handle        | version | solutionName  | solutionDesc           | workflowName     | workflowHandle | workflowDesc     |
      | my-parent | com.genaiz.test | my-solution-2 | 1.0.0   | My Solution 2 | Solution 2 Description | Default Workflow | default        | default workflow |
    And the folder "<folder>" created
    And the workdir changed to "<folder>"
    When I run the command "sn create <handle> --oem=<oem> --name='<solutionName>' --description='<solutionDesc>'"
    Then I should have a solution under "<folder>/<handle>" named "<solutionName>" with oem "<oem>", handle "<folder>", description "<solutionDesc>" and version "<version>"
    And I should have a workflow under "<folder>/<handle>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: list simple solution with dot, recursively
    Given the scenario "create simple solution with name and description" ran with condition "service_completed_successfully"
    And the following parameters
      | oem            | handle        | version | solutionName  | local |
      | com.genaiz.test | my-solution-2 | 1.0.0   | My Solution 2 | true  |
    When I run the command "sn list . --json"
    Then I should have a solution list with solution "<oem>/<handle>:<version>" named "<solutionName>" and local flag set to "<local>"
