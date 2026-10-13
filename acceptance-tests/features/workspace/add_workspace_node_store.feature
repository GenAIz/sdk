Feature: add data store to workspace node
  To be able to add a data store to a workspace node
  As an authenticated user,
  I need to be able to create a solution with a connector function, and add the function to the solution default workflow.
  I need to be able to create a datalink, add properties to it, login to an orchestrator and publish it.
  I need to be able to add a data store to the function with the datalink and publish the solution
  I need to be able to create a locker, add the datalink to it as data store, update the properties to test values, and publish it.
  I need to be able to create a workspace with the workflow of the solution, list the single node and add the data store to it.

  Scenario: create workspace solution
    Given the following parameters
      | handle      | oem             | version | workflowDesc     | workflowHandle | workflowName     |
      | my-solution | com.genaiz.test | 1.0.0   | default workflow | default        | Default Workflow |
    When I run the command "sn create <handle> --oem=<oem>"
    Then I should have a solution under "<handle>" named "<handle>" with oem "<oem>", handle "<handle>", description "<handle>" and version "<version>"
    And I should have a workflow under "<handle>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create solution function
    Given the scenario "create workspace solution" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | oem             | handle      | version | recipe       | type      |
      | my-solution | com.genaiz.test | my-function | 1.0.0   | bash-example | connector |
    And the workdir changes to "<path>"
    When I run the command "sf create <handle> --type=<type>"
    Then I should have a function under "<path>/<handle>" named "<handle>" with oem "<oem>", handle "<handle>", version "<version>" and type "<type>"

  Scenario: add node to default solution workflow
    Given the scenario "create solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | functionPath | workflowHandle |
      | my-solution | my-function  | default        |
    And the workdir changes to "<path>"
    When I run the command "wf nodes add <workflowHandle> <functionPath>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionHandle>" and version "<version>"

  Scenario: create datalink for function
    Given the following parameters
      | path                  | configFile  | handle     | oem             | version |
      | $HOME/.config/genaiz/ | Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add datalink property for function
    Given the scenario "create datalink for function" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem             | version | key        | type   | defaultValue |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   | IP_ADDRESS | STRING | 192.168.1.1  |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --default-value=<defaultValue>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

  Scenario: login to orchestrator
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                      | username         |
      | success  | $HOME/.cache/genaiz/.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish datalink for function
    Given the scenario "login to orchestrator" ran with condition "service_completed_successfully"
    And the scenario "add datalink property for function" ran with condition "service_completed_successfully"
    And the following parameters
      | handle     | oem             | version |
      | datalink-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: add datalink for function as data store
    Given the scenario "publish datalink for function" ran with condition "service_completed_successfully"
    And the scenario "create solution function" ran with condition "service_completed_successfully"
    And the following parameters
      | path                    | oem             | handle     | version |
      | my-solution/my-function | com.genaiz.test | datalink-1 | 1.0.0   |
    And the workdir changes to "<path>"
    When I run the command "sf data str add <oem>/<handle>:<version>"
    Then I should have a data store under "<path>" with datalink "<oem>/<handle>:<version>"

  Scenario: build function for solution
    Given the scenario "add datalink for function as data store" ran with condition "service_completed_successfully"
    And the following parameters
      | solution    | folder      | oem             |
      | my-solution | my-function | com.genaiz.test |
    And the workdir changes to "<solution>/<folder>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: publish solution to orchestrator
    Given the scenario "build function for solution" ran with condition "service_completed_successfully"
    And the scenario "add node to default solution workflow" ran with condition "service_completed_successfully"
    And the following parameters
      | path        | functionHandle |
      | my-solution | my-function    |
    And the workdir changes to "<path>"
    When I run the command "sn publish"
    Then I should have a published solution "<path>" with a smart function "<functionHandle>"

  Scenario: create a data locker
    Given the following parameters
      | path         | password  |
      | myLocker.bin | data$T0rE |
    And the environment contains "GENAIZ_LK_PASSWORD='<password>'"
    When I run the command "lk init <path>"
    Then I should have a locker file under "<path>"

  Scenario: add data store to data locker
    Given the scenario "create a data locker" ran with condition "service_completed_successfully"
    And the scenario "publish datalink for function" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password  | handle      | dataLinkFqdn               | dataLinkVersion | mtime |
      | myLocker.bin | data$T0rE | myLockerStr | com.genaiz.test/datalink-1 | 1.0.0           |       |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD='<password>'"
    When I run the command "lk str add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data store property
    Given the scenario "add data store to data locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password  | handle      | key        | value         |
      | myLocker.bin | data$T0rE | myLockerStr | IP_ADDRESS | 192.168.1.101 |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD='<password>'"
    When I run the command "lk str update <handle> <key> <value> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: publish data store to account
    Given the scenario "update data store property" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password  | handle      |
      | myLocker.bin | data$T0rE | myLockerStr |
    And the environment contains "GENAIZ_LK_PASSWORD='<password>'"
    When I run the command "lk str publish <handle> --locker=<path>"
    Then I should have a data store named "<handle>" created under account "<orchestrator>"

  Scenario: create workspace for solution
    Given the scenario "login to orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | name                  | visibility | flags |
      | my-solution-workspace | PRIVATE    | 3     |
    When I run the command "ws create <name> --json"
    Then I should have a workspace with name "<name>", a created timestamp, the visibility set to "<visibility>" and flags set to "<flags>"

  Scenario: create workspace flow for solution
    Given the scenario "create workspace for solution" ran with condition "service_completed_successfully"
    And the scenario "publish solution to orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName         | oem             | solutionHandle | solutionVersion | wfHandle |
      | my-solution-workspace | com.genaiz.test | my-solution    | 1.0.0           | default  |
    When I run the command "ws flow create <workspaceName> <oem>/<solutionHandle>:<solutionVersion> <wfHandle> --json"
    Then I should have a workspace flow for workflow "<wfHandle>" and solution "<oem>/<solutionHandle>:<solutionVersion>"

  Scenario: list workspace flow node of solution
    Given the scenario "create workspace flow for solution" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName         | wfHandle | nodeHandle       |
      | my-solution-workspace | default  | my-function-node |
    When I run the command "ws node list <workspaceName> <wfHandle> --json"
    Then I should have a list of nodes with a node named "<nodeHandle>" and handle "<nodeHandle>"

  Scenario: attach data store to workspace flow node
    Given the scenario "create workspace flow for solution" ran with condition "service_completed_successfully"
    And the scenario "publish data store to account" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName         | workflowHandle | nodeHandle       | dsName      |
      | my-solution-workspace | default        | my-function-node | myLockerStr |
    When I run the command "ws node data str add <workspaceName> <workflowHandle> <nodeHandle> <dsName>"
    Then I should have a data store named "<dsName>" attached to workspace "<workspaceName>" on flow node "<workflowHandle>/<nodeHandle>"

  Scenario: detach data store to workspace flow node
    Given the scenario "attach data store to workspace flow node" ran with condition "service_completed_successfully"
    And the following parameters
      | workspaceName         | workflowHandle | nodeHandle       | dsName      |
      | my-solution-workspace | default        | my-function-node | myLockerStr |
    When I run the command "ws node data str rm <workspaceName> <workflowHandle> <nodeHandle> <dsName>"
    Then I should not have any data stores on flow node "<workflowHandle>/<nodeHandle>"
