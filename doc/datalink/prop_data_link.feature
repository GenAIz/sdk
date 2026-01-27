Feature: data link properties
  To be able to manage property specifications of a data link
  As a local user
  I should be able to create a data link, add, edit and remove property and secret property specifications from the created data link

  Scenario: create data link
    Given the following parameters
      | configFile                       | handle     | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link property
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | TEST_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --type=<type>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value ""

  Scenario: edit data link property
    Given the scenario "add data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      | type   | defaultValue |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | TEST_KEY | STRING | def_val      |
    When I run the command "dk prop edit <oem>/<handle>:<version> <key> --default-value=<defaultValue>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value "<defaultValue>"

  Scenario: add data link secret property
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key        |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | SECRET_KEY |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: edit data link secret property invalid default value
    Given the scenario "add data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key        | defaultValue  |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | SECRET_KEY | invalid_value |
    When I run the command "dk prop edit <oem>/<handle>:<version> <key> --default-value=<defaultValue>"
    Then I should have an error for key "datalink.propspecedit.defaultvalue" with value "<defaultValue>"

  Scenario: edit data link secret property
    Given the scenario "add data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key        | name | description |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | SECRET_KEY | name | description |
    When I run the command "dk prop edit <oem>/<handle>:<version> <key> --name=<name> --description=<description>"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>", name "<name>", and description "<description>"

  Scenario: remove data link property
    Given the scenario "edit data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | TEST_KEY |
    When I run the command "dk prop rm <oem>/<handle>:<version> <key>"
    Then I should not have a property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", for key "<key>"

  Scenario: remove data link secret property
    Given the scenario "remove data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | key      |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | SECRET_KEY |
    When I run the command "dk prop rm <oem>/<handle>:<version> <key>"
    Then I should not have a secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", for key "<key>"
