Feature: data sources for the bash example
  To be able to add and remove data sources to the bash example function
  As an authenticated developer
  I should be able to create the bash example recipe, add data links, add data sources and remove one from the function

  Scenario: create bash example
    Given the following parameters
      | recipe       | handle          | oem            | type     | version |
      | bash-example | my-bash-example | com.genaiz.dev | function | 1.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --version=<version>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: login bash example
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: create data link default version
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<folder>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: create data link additional
    Given the scenario "create data link default version" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-2 | com.genaiz.dev | 2.0.0   |
    When I run the command "dk create <handle> <folder> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<folder>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data source invalid handle
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     |
      | my-bash-example | not--valid |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <handle>"
    Then I should have an error for field "function.publish.datasourceadd.handle"

  Scenario: add data source invalid oem
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem        |
      | my-bash-example | datalink-2 | .not.valid |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>"
    Then I should have an error for field "function.publish.datasourceadd.oem"

  Scenario: add data source invalid version
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-2 | com.genaiz.dev | ..1     |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have an error for field "function.publish.datasourceadd.version"

  Scenario: add data source wrong version
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version | available |
      | my-bash-example | datalink-2 | com.genaiz.dev | 1.0.0   | 2.0.0     |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle> --version=<version>"
    Then I should have a data source unavailable error with version "<available>" listed as available

  Scenario: add data source additional
    Given the scenario "create data link additional" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-2 | com.genaiz.dev | 2.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <oem>/<handle>:<version>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario: add data source one
    Given the scenario "add data source additional" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source add <handle>:<version> --oem=<oem>"
    Then I should have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"

  Scenario remove data source one
    Given the scenario "add data source one" ran with condition "service_completed_successfully"
    And the following parameters
      | folder          | handle     | oem            | version |
      | my-bash-example | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the workdir changes to "<folder>"
    When I run the command "sf data source rm <handle>:<version> --oem=<oem>"
    Then I should not have a data source under "<folder>" with datalink "<oem>/<handle>:<version>"
