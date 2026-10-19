Feature: set workspace node input file
  To be able to set an input file on a workspace node
  As an authenticated user,
  I need to be able to create a solution with a function, and add the function to the solution default workflow.
  I need to be able to be able to add a data input port to the function, and add it as a workflow node.
  I need to be able to login to an orchestrator and publish the solution.
  I need to be able to create a workspace with the workflow of the solution, list the single function node and add a data set to it.
  I need to be able to list the data set from the assigned node.

  Scenario: create workspace solution
    Given the following parameters
      | handle      | oem             | version | workflowHandle |
      | my-solution | com.genaiz.test | 1.0.0   | my-workflow    |
    When I run the command "sn create <handle> --oem=<oem> --workflow-handle=<workflowHandle>"
    Then I should have a solution under "<handle>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<handle>" named "<workflowHandle>", handle "<workflowHandle>" with description "<workflowHandle>"

  Scenario: create solution function
    Given the scenario "create workspace solution" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | oem             | handle      | version | recipe       | type     |
      | my-solution | com.genaiz.test | my-function | 1.0.0   | bash-example | function |
    And the workdir changes to "<path>"
    When I run the command "sf create <handle>"
    Then I should have a function under "<path>/<handle>" named "<handle>" with oem "<oem>", handle "<handle>", version "<version>" and type "<type>"

  Scenario: add input data port to function
    Given the scenario "create solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | path                    | inputHandle |
      | my-solution/my-function | configs     |
    And the workdir changes to "<path>"
    When I run the command "sf data input add <inputHandle>"
    Then I should have an input port under "<path>" with handle "<inputHandle>", named "<inputHandle>" with description ""

  Scenario: add node to solution workflow
    Given the scenario "create solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | functionPath | workflowHandle |
      | my-solution | my-function  | my-workflow    |
    And the workdir changes to "<path>"
    When I run the command "wf nodes add <workflowHandle> <functionPath>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionHandle>" and version "<version>"

  Scenario: login to orchestrator
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: build function for solution
    Given the scenario "add input data port to function" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | oem             |
      | my-solution | my-function | com.genaiz.test |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: publish solution to orchestrator
    Given the scenario "build function for solution" ran with condition "service_completed_successfully"
    And the scenario "add node to solution workflow" ran with condition "service_completed_successfully"
    And the scenario "login to orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | functionHandle |
      | my-solution | my-function    |
    And the workdir changes to "<path>"
    When I run the command "sn publish"
    Then I should have a published solution "<path>" with a smart function "<functionHandle>"

  Scenario: create workspace for solution
    Given the scenario "login to orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | name         | visibility | flags |
      | my-workspace | PRIVATE    | 3     |
    When I run the command "ws create <name> --json"
    Then I should have a workspace with name "<name>", a created timestamp, the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace flow for solution
    Given the scenario "create workspace for solution" ran with condition "service_completed_successfully"
    And the scenario "publish solution to orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName | oem             | solutionHandle | solutionVersion | wfHandle    |
      | my-workspace  | com.genaiz.test | my-solution    | 1.0.0           | my-workflow |
    When I run the command "ws flow create <workspaceName> <oem>/<solutionHandle>:<solutionVersion> <wfHandle> --json"
    Then I should have a workspace flow for workflow "<wfHandle>" and solution "<oem>/<solutionHandle>:<solutionVersion>"

  Scenario: list workspace flow node of solution
    Given the scenario "create workspace flow for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName | wfHandle    | nodeHandle       |
      | my-workspace  | my-workflow | my-function-node |
    When I run the command "ws node list <workspaceName> <wfHandle> --json"
    Then I should have a list of nodes with a node named "<nodeHandle>" and handle "<nodeHandle>"

  Scenario: create configuration file for data set
    Given the following parameters
      | path                    | file        |
      | my-solution/my-function | config.json |
    And the workdir changes to "<path>"
    # We will need a different docker image to run the command or a tweak of the entrypoint
    When I run the command "echo '{}' | tee <file>"
    Then I should have a non-empty file "<file>" under "<path>"

  Scenario: add data set to workspace flow node
    Given the scenario "create configuration file for data set" ran with condition "service_completed_successfully"
    And the scenario "list workspace flow node of solution" ran with condition "service_completed_successfully"
    And the following parameters
      | path                    | file        | port    | workspaceName | workflowHandle | sfHandle    | portHandle | setHandle  |
      | my-solution/my-function | config.json | configs | my-workspace  | my-workflow    | my-function | configs    | my-dataset |
    When I run the command "ws node data set <workspaceName> <workflowHandle> <sfHandle> <portHandle> <path>/<file> --set-handle=<setHandle> --json"
    Then I should have a data-set with handle "<setHandle>" and file "<file>"

  Scenario: list data set of workspace flow node
    Given the scenario "add data set to workspace flow node" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName | workflowHandle | sfHandle    | setHandle  | portHandle |
      | my-workspace  | my-workflow    | my-function | my-dataset | configs    |
    When I run the command "ws node data list <workspaceName> <workflowHandle> <sfHandle>"
    Then I should have a data-set list with an entry for "<setHandle>" associated with "<portHandle>"

