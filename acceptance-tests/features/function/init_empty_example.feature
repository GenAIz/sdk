Feature: function init on empty folder
  To be able to initialize a function
  As a developer
  I should be able to invoke init in an empty folder

  Scenario: init empty example bad context
    Given the following parameters
      | context     |
      | /badContext |
    When I run the command "sf init --context=<context>"
    Then I should have an error for field "function.build.context"

  Scenario: init empty example bad Dockerfile
    Given the following parameters
      | file          |
      | badDockerfile |
    When I run the command "sf init --file=<file>"
    Then I should have an error for field "function.build.context"

  Scenario: init empty example bad config-type
    Given the following parameters
      | configType |
      | invalid    |
    When I run the command "sf init --config-type=<configType>"
    Then I should have an error for field "function.init.configtype"

  Scenario: init empty example bad repository
    Given the following parameters
      | repository                     |
      | com.genaiz.test/bad..repository |
    When I run the command "sf init --repository=<repository>"
    Then I should have an error for field "function.build.repository"

  Scenario: init empty example bad handle
    Given the following parameters
      | handle         |
      | empty__example |
    When I run the command "sf init --handle=<handle>"
    Then I should have an error for field "function.init.handle"

  Scenario: init empty example bad oem
    Given the following parameters
      | handle        | oem         |
      | empty-example | com..genaiz |
    When I run the command "sf init --handle=<handle> --oem=<oem>"
    Then I should have an error for field "function.init.oem"

  Scenario: init empty example bad version
    Given the following parameters
      | handle        | oem            | version |
      | empty-example | com.genaiz.test | 1..0    |
    When I run the command "sf init --handle=<handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "function.init.version"

  Scenario: init empty example bad type
    Given the following parameters
      | handle        | oem            | version | type    |
      | empty-example | com.genaiz.test | 1.0.1   | invalid |
    When I run the command "sf init --handle=<handle> --oem=<oem> --version=<version> --type=<type>"
    Then I should have an error for field "function.init.type"

  Scenario: init empty example bad name
    Given the following parameters
      | handle        | name                                                                                                                                                                                                                                                                                            |
      | empty-example | This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. |
    When I run the command "sf init --handle=<handle> --name='<name>'"
    Then I should have an error for field "function.init.name"

  Scenario: init empty example bad arch
    Given the following parameters
      | folder        | arch  |
      | empty-example | amd37 |
    When I run the command "sf init --handle=<handle> --arch=<arch>"
    Then I should have an error for field "function.init.arches"

  Scenario: create empty example solution
    Given the following parameters
      | path              | oem            | version |
      | my-empty-solution | com.genaiz.test | 0.1.1   |
    When I run the command "sn create <path> --oem=<oem> --version=<version>"
    Then I should have a solution under "<path>" named "<path>" with version "<version>"

  Scenario: init empty example
    Given the scenario "create empty example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | handle            | oem            | version | type     |
      | my-empty-solution | my-empty-function | com.genaiz.test | 0.1.1   | function |
    And the working dir <handle> created
    And a Dockerfile created under <handle>
    When I run the command "sf init"
    Then I should have a function under "<handle>" named "<handle>", with handle "<handle>", oem "<oem>", version "<version>" and type "<type>"
