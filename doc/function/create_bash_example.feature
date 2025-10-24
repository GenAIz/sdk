Feature: function create for the bash example
  To be able to create the bash example function
  As a developer
  I should be able to create the bash example recipe

  Scenario: create bash example bad context
    Given the following parameters
      | handle          | context         |
      | my-bash-example | /_badContext |
    When I run the command "sf create <handle> --context=<context>"
    Then I should have an error for field "sf.build.context"

  Scenario: create bash example bad config-type
    Given the following parameters
      | handle          | configType |
      | my-bash-example | invalid    |
    When I run the command "sf create <handle> --config-type=<configType>"
    Then I should have an error for field "sf.create.configtype"

  Scenario: create bash example bad tag
    Given the following parameters
      | handle          | tag                     |
      | my-bash-example | com.genaiz.dev/bad..tag |
    When I run the command "sf create <handle> --tag=<tag>"
    Then I should have an error for field "sf.build.tag"

  Scenario: create bash example bad handle
    Given the following parameters
      | handle        |
      | bash__example |
    When I run the command "sf create <handle>"
    Then I should have an error for field "sf.create.handle"

  Scenario: create bash example bad oem
    Given the following parameters
      | handle       | oem         |
      | bash-example | com..genaiz |
    When I run the command "sf create <handle> --oem=<oem>"
    Then I should have an error for field "sf.create.oem"

  Scenario: create bash example bad version
    Given the following parameters
      | handle       | oem            | version |
      | bash-example | com.genaiz.dev | 1..0    |
    When I run the command "sf create <handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "sf.create.version"

  Scenario: create bash example bad recipe
    Given the following parameters
      | handle       | recipe     |
      | bash-example | bad-recipe |
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have an error for field "sf.create.recipe"

  Scenario: create bash example bad type
    Given the following parameters
      | handle       | oem            | version | type    |
      | bash-example | com.genaiz.dev | 1.0.0   | invalid |
    When I run the command "sf create <handle> --oem=<oem> --version=<version> --type=<type>"
    Then I should have an error for field "sf.create.type"

  Scenario: create bash example bad name
    Given the following parameters
      | handle       | name                                                                                                                                                                                                                                                                                            |
      | bash-example | This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. |
    When I run the command "sf create <handle> --name='<name>'"
    Then I should have an error for field "sf.create.name"

  Scenario: create bash example bad arch
    Given the following parameters
      | handle       | arch  |
      | bash-example | amd37 |
    When I run the command "sf create <handle> --arch=<arch>"
    Then I should have an error for field "sf.create.arches"

  Scenario: create bash example solution
    Given the following parameters
      | path             | oem            | version |
      | my-bash-solution | com.genaiz.dev | 0.1.1   |
    When I run the command "sn create <path> --oem=<oem> --version=<version>"
    Then I should have a solution under "<path>" named "<path>" with version "<version>"

  Scenario: create bash example oem default values
    Given the scenario "create bash example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | recipe       | handle          | oem            | type     | version |
      | bash-example | my-bash-example | com.genaiz.dev | function | 0.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"
