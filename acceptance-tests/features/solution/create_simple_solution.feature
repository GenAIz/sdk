Feature: solution create for a simple solution
  To be able to create a simple solution
  As a developer
  I should be able to create a solution

  Scenario: create simple solution bad config-type
    Given the following parameters
      | folder      | configType |
      | my-solution | invalid    |
    When I run the command "sn create <folder> --config-type=<configType>"
    Then I should have an error for field "solution.create.configtype"

  Scenario: create simple solution bad handle
    Given the following parameters
      | folder      | handle    |
      | my-solution | invalid.. |
    When I run the command "sn create <folder> --handle=<handle>"
    Then I should have an error for field "solution.create.handle"

  Scenario: create simple solution bad oem
    Given the following parameters
      | folder      | oem       |
      | my-solution | invalid.. |
    When I run the command "sn create <folder> --oem=<oem>"
    Then I should have an error for field "solution.create.oem"

  Scenario: create simple solution bad name
    Given the following parameters
      | folder      | name                                                                                                                                                                                                                                                                |
      | my-solution | this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long because |
    When I run the command "sn create <folder> --name='<name>'"
    Then I should have an error for field "solution.create.name"

  Scenario: create simple solution bad version
    Given the following parameters
      | folder      | version       |
      | my-solution | 1.0.0-invalid |
    When I run the command "sn create <folder> --version=<version>"
    Then I should have an error for field "solution.create.oem"

  Scenario: create simple solution bad workflow handle
    Given the following parameters
      | folder      | workflowHandle |
      | my-solution | invalid..      |
    When I run the command "sn create <folder> --workflow-handle=<workflowHandle>"
    Then I should have an error for field "solution.create.workflow.handle"

  Scenario: create simple solution bad workflow name
    Given the following parameters
      | folder      | workflowName                                                                                                                                                                                                                                                        |
      | my-solution | this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long this name value is too long because |
    When I run the command "sn create <folder> --workflow-name='<workflowName>'"
    Then I should have an error for field "solution.create.workflow.name"

  Scenario: create simple solution
    Given the following parameters
      | folder      | oem            | handle     | description          | name        | version | workflowDesc         | workflowHandle | workflowName |
      | my-solution | dev.genaiz.com | solution-1 | solution description | My Solution | 0.1.1   | workflow description | workflow-1     | workflow one |
    When I run the command "sn create <folder> --oem=<oem> --handle=<handle> --description='<description>' --name='<name>' --version=<version> --workflow-desc='<workflowDesc>' --workflowHandle='<workflowHandle>' --workflowName='<workflowName>'"
    Then I should have a solution under "<folder>" named "<name>" with oem "<oem>", handle "<handle>", description "<description>" and version "<version>"
    And I should have a workflow under "<folder>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"
