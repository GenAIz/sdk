Feature: workflow delete for a simple solution
  To be able to delete a workflow
  As a developer
  I should be able to create a solution, create a new workflow, then delete that workflow

  Scenario: create simple solution with workflow
    Given the following parameters
      | folder      | oem            | version | workflowHandle | workflowName     | workflowDescription |
      | my-solution | com.genaiz.dev | 1.1.0   | my-workflow    | Default Workflow | default workflow    |
    When I run the command "sn create <folder> --oem=<oem> --version=<version> --workflow-handle=<workflowHandle>"
    Then I should have a solution under "<folder>" named "<folder>" with oem "<oem>", handle "<folder>", description "<folder>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDescription>"

  Scenario: create simple workflow
    Given the scenario "create simple solution with workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle |
      | my-solution | simpleWorkflow |
    When I run the command "wf create <workflowHandle> <folder>"
    Then I should have a workflow under "<folder>" named "<workflowHandle>", handle "<workflowHandle>" with description "<workflowHandle>"

  Scenario: delete simple workflow
    Given the scenario "create simple workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | folder      | workflowHandle |
      | my-solution | simpleWorkflow |
    And the workdir changes to "<folder>"
    When I run the command "wf delete <workflowHandle>"
    Then I should not have a workflow under "<folder>" with handle "<workflowHandle>"