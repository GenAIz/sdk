Feature: data link outbound proxies
  To be able to manage outbound proxy configurations for a data link
  As a local user
  I should be able to create a data link, add  and remove outbound proxy configurations for the created data link

  Scenario: create data link
    Given the following parameters
      | configFile                       | handle     | oem            | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add portless data link proxy
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | address        |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | dev.genaiz.com |
    When I run the command "dk proxy add <oem>/<handle>:<version> <address>"
    Then I should have an error "Error: address <address>: missing port in address"

  Scenario: add data link proxy
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | address        | port |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | *              |      |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | dev.genaiz.com | :443 |
    When I run the command "dk proxy add <oem>/<handle>:<version> <address><port>"
    Then I should have an outbound proxy for data link "<oem>/<handle>:<version>" for full address "<address><port>"

  Scenario: remove invalid data link proxy
    Given the scenario "add data link proxy" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle           | oem            | version | address        | port |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-invalid | com.genaiz.dev | 1.0.0   | dev.genaiz.com | :443 |
    When I run the command "dk proxy rm <oem>/<handle>:<version> <address><port>"
    Then I should have an error "Error: data link [<oem>/<handle>:<version>] not found"

  Scenario: remove data link proxy
    Given the scenario "add data link proxy" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem            | version | address        | port |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.dev | 1.0.0   | dev.genaiz.com | :443 |
    When I run the command "dk proxy rm <oem>/<handle>:<version> <address><port>"
    Then I should not have an outbound proxy for data link "<oem>/<handle>:<version>"
