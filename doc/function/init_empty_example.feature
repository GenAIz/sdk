Feature: function init on empty folder
  To be able to initialize a function
  As a developer
  I should be able to invoke init in an empty folder

  Scenario: init empty example bad context
    Given the following parameters
      | context     |
      | /badContext |
    When I run the command "sf init --context=<context>"
    Then I should have an error for field "sf.build.context"

  Scenario: init empty example bad Dockerfile
    Given the following parameters
      | file          |
      | badDockerfile |
    When I run the command "sf init --file=<file>"
    Then I should have an error for field "sf.build.context"

  Scenario: init empty example bad config-type
    Given the following parameters
      | configType |
      | invalid    |
    When I run the command "sf init --config-type=<configType>"
    Then I should have an error for field "sf.init.configtype"

  Scenario: init empty example bad tag
    Given the following parameters
      | tag                     |
      | com.genaiz.dev/bad..tag |
    When I run the command "sf init --tag=<tag>"
    Then I should have an error for field "sf.build.tag"

  Scenario: init empty example bad handle
    Given the following parameters
      | handle         |
      | empty__example |
    When I run the command "sf init --handle=<handle>"
    Then I should have an error for field "sf.init.handle"

  Scenario: init empty example bad oem
    Given the following parameters
      | handle        | oem         |
      | empty-example | com..genaiz |
    When I run the command "sf init --handle=<handle> --oem=<oem>"
    Then I should have an error for field "sf.init.oem"

  Scenario: init empty example bad version
    Given the following parameters
      | handle        | oem            | version |
      | empty-example | com.genaiz.dev | 1..0    |
    When I run the command "sf init --handle=<handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "sf.init.version"

  Scenario: init empty example bad type
    Given the following parameters
      | handle        | oem            | version | type    |
      | empty-example | com.genaiz.dev | 1.0.0   | invalid |
    When I run the command "sf init --handle=<handle> --oem=<oem> --version=<version> --type=<type>"
    Then I should have an error for field "sf.init.type"

  Scenario: init empty example bad name
    Given the following parameters
      | handle        | name                                                                                                                                                                                                                                                                                            |
      | empty-example | This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. |
    When I run the command "sf init --handle=<handle> --name='<name>'"
    Then I should have an error for field "sf.init.name"

  Scenario: init empty example bad arch
    Given the following parameters
      | folder        | arch  |
      | empty-example | amd37 |
    When I run the command "sf init --handle=<handle> --arch=<arch>"
    Then I should have an error for field "sf.init.arches"

  Scenario: create empty example solution
    Given the following parameters
      | path              | oem            | version |
      | my-empty-solution | com.genaiz.dev | 0.1.1   |
    When I run the command "sn create <path> --oem=<oem> --version=<version>"
    Then I should have a solution under "<path>" named "<path>" with version "<version>"

  Scenario: init empty example
    Given the scenario "create empty example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | path              | handle            | oem            | version |
      | my-empty-solution | my-empty-function | com.genaiz.dev | 0.1.1   |
    And the working dir <handle> created
    When I run the command "sf init"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
