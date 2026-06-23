Feature: function create for the bash example
  To be able to create the bash example function
  As a developer
  I should be able to create the bash example recipe

  Scenario: create bash example bad context
    Given the following parameters
      | handle          | context      |
      | my-bash-example | /_badContext |
    When I run the command "sf create <handle> --context=<context>"
    Then I should have an error for field "function.build.context"

  Scenario: create bash example bad config-type
    Given the following parameters
      | handle          | configType |
      | my-bash-example | invalid    |
    When I run the command "sf create <handle> --config-type=<configType>"
    Then I should have an error for field "function.create.configtype"

  Scenario: create bash example bad repository
    Given the following parameters
      | handle          | repository               |
      | my-bash-example | com.genaiz.test/bad..repo |
    When I run the command "sf create <handle> --repository=<repository>"
    Then I should have an error for field "function.build.repository"

  Scenario: create bash example bad handle
    Given the following parameters
      | handle        |
      | bash__example |
    When I run the command "sf create <handle>"
    Then I should have an error for field "function.create.handle"

  Scenario: create bash example bad oem
    Given the following parameters
      | handle       | oem         |
      | bash-example | com..genaiz |
    When I run the command "sf create <handle> --oem=<oem>"
    Then I should have an error for field "function.create.oem"

  Scenario: create bash example bad version
    Given the following parameters
      | handle       | oem            | version |
      | bash-example | com.genaiz.test | 1..0    |
    When I run the command "sf create <handle> --oem=<oem> --version=<version>"
    Then I should have an error for field "function.create.version"

  Scenario: create bash example bad recipe
    Given the following parameters
      | handle       | recipe     |
      | bash-example | bad-recipe |
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have an error for field "function.create.recipe"

  Scenario: create bash example bad type
    Given the following parameters
      | handle       | oem            | version | type    |
      | bash-example | com.genaiz.test | 1.0.1   | invalid |
    When I run the command "sf create <handle> --oem=<oem> --version=<version> --type=<type>"
    Then I should have an error for field "function.create.type"

  Scenario: create bash example bad name
    Given the following parameters
      | handle       | name                                                                                                                                                                                                                                                                                            |
      | bash-example | This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. This string is too long for a name. |
    When I run the command "sf create <handle> --name='<name>'"
    Then I should have an error for field "function.create.name"

  Scenario: create bash example bad arch
    Given the following parameters
      | handle       | arch  |
      | bash-example | amd37 |
    When I run the command "sf create <handle> --arch=<arch>"
    Then I should have an error for field "function.create.arches"

  Scenario: create bash example solution
    Given the following parameters
      | path             | oem            | version |
      | my-bash-solution | com.genaiz.test | 0.1.1   |
    When I run the command "sn create <path> --oem=<oem> --version=<version>"
    Then I should have a solution under "<path>" named "<path>" with version "<version>"

  Scenario: create bash example oem default values
    Given the scenario "create bash example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | recipe       | handle          | oem            | type     | version |
      | bash-example | my-bash-example | com.genaiz.test | function | 0.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario create bash example with mounts
    Given the scenario "create bash example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | recipe       | handle         | oem            | version | name        | mountPoint | mountIn     | mountOut                 | mountLog                 | mountVar                 |
      | bash-example | my-bash-mounts | com.genaiz.test | 0.0.1   | Bash Mounts | run/test   | run/test/in | run/test/{timestamp}/out | run/test/{timestamp}/log | run/test/{timestamp}/var |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --name='<name>' --mount-in=<mountIn> --mount-out=<mountOut>
    Then I should have a function under "<handle>" named "<name>" with oem "<oem>", version "<version>" and type "<type>"
    And I should have a function under "<handle>" with run mount input "<mountIn>" with the folder created under "<handle>/<mountIn>"
    And I should have a function under "<handle>" with run mount output "<mountOut>" with a stamped folder under "<handle>/<mountPoint>"
    And I should have a function under "<handle>" with run mount log "<mountLog>" with a stamped folder under "<handle>/<mountPoint>"
    And I should have a function under "<handle>" with run mount var "<mountVar>" with a stamped folder under "<handle>/<mountPoint>"
