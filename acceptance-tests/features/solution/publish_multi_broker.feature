Feature: solution publish, using multiple brokers
  To be able to publish a simple solution to multiple brokers
  As an authenticated user,
  I should be able to create a solution, create a function, add it as node of the solution, and publish the solution to an orchestrator
  I should be able to login onto a second orchestrator, modify the function published, and publish it to this second orchestrator.
  I should be able to re-publish the modified function to the first orchestrator

  Scenario: create simple solution
    Given the following parameters
      | handle         | oem             | version | name                  | workflowName     | workflowHandle | workflowDescription |
      | multi-solution | com.genaiz.test | 1.0.0   | Multi Broker Solution | Default Workflow | default        | default workflow    |
    When I run the command "sn create <folder> --name='<name>' --oem=<oem>"
    Then I should have a solution under "<handle>" named "<name>" with oem "<oem>", handle "<handle>", description "<name>" and version "<version>"
    And I should have a workflow under "<handle>" named "<workflowName>", handle "<workflowHandle>" with description "<workflowDesc>"

  Scenario: create bash example
    Given the scenario "create simple solution" ran with condition "service_completed_successfully"
    And the following parameters
      | folder         | recipe       | handle        | oem             | type     | version |
      | multi-solution | bash-example | multi-example | com.genaiz.test | function | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: build bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | folder         | handle        | oem             |
      | multi-solution | multi-example | com.genaiz.test |
    And the workdir changes to "<folder>/<handle>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: add bash example workflow node
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder         | functionHandle | workflowHandle | nodeHandle | oem             | version |
      | multi-solution | multi-example  | default        | my-node    | com.genaiz.test | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "wf nodes add <workflowHandle> <nodeHandle> --sf=<oem>/<functionHandle>:<version>"
    Then I should have a node under "<folder>" and workflow "<workflowHandle> named "<nodeHandle>" and handle "<nodeHandle>"
    And I should have a smart function under "<folder>", workflow "<workflowHandle>", node "<nodeHandle>" with oem "<oem>", handle "<functionFolder>" and version "<version>"

  Scenario: login first orchestrator
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish simple solution
    Given the scenario "login first orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | solutionHandle | functionHandle |
      | multi-solution | multi-example  |
    And the workdir changes to "<solutionHandle>"
    When I run the command "sn publish"
    Then I should have a published solution "<solutionHandle>" with a smart function "<functionHandle>"

  Scenario: modify bash example
    Given the scenario "publish bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder                       | src    |
      | multi-solution/multi-example | app.sh |
    And the workdir changes to "<folder>"
    # will need a way to specify another image with a different entry point, or we'll need sed to be bundled and use a different entrypoint
    When I run the command "sed -i 's/sleep 4s/sleep 5s/' <src>"
    Then I should have a modified "<folder>/<src>" file

  Scenario: build modified bash example
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the execution group "<docker_gid>"
    And the following parameters
      | handle        | oem             |
      | multi-example | com.genaiz.test |
    And the workdir changes to "<handle>"
    When I run the command "sf build"
    Then I should have a docker image tagged "<oem>/<handle>:latest"

  Scenario: login second orchestrator
    Given the second orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <secondOrchestrator> --username=<username>"
    Then I should have an active session id with host "<secondOrchestrator>" for username "<username>" under path "<path>"

  Scenario: publish simple solution to second orchestrator
    Given the scenario "login first orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | solutionHandle | functionHandle |
      | multi-solution | multi-example  |
    And the workdir changes to "<solutionHandle>"
    When I run the command "sn publish"
    Then I should have a published solution "<solutionHandle>" with a smart function "<functionHandle>"

  Scenario: activate orchestrator session
    Given the second orchestrator is running with condition: "service_healthy"
    And the following parameters
      | path                     |
      | $HOME/.cache/genaiz.auth |
    When I run the command "ac activate <orchestrator>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish modified solution
    Given the scenario "login first orchestrator" ran with condition "service_completed_successfully"
    And the following parameters
      | solutionHandle | functionHandle |
      | multi-solution | multi-example  |
    And the workdir changes to "<solutionHandle>"
    When I run the command "sn publish"
    Then I should have a published solution "<solutionHandle>" with a smart function "<functionHandle>"
