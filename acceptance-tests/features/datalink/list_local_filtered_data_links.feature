Feature: list local filtered data links
  To be able to list data links
  As an authenticated developer
  I should be able to create 2 data links in different oem folders and list them filtered by oem afterwards

  Scenario: create first data link
    Given the following parameters
      | configFolder | configFile           | handle     | oem             | version |
      | firstOem     | firstOem/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> <configFolder> --oem=<oem>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: create second data link
    Given the following parameters
      | configFolder | configFile            | handle     | oem             | version |
      | secondOem    | secondOem/Genaiz.yaml | datalink-2 | com.genaiz.test | 1.0.1   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> <configFolder> --oem=<oem> --version=<version>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: list first data link
    Given the scenario "create first data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFolder | name       | fqdn                      | notFqdn                    | notName    |
      | firstOem     | datalink-1 | firstOem/datalink-1:1.0.0 | secondOem/datalink-2:1.0.1 | datalink-2 |
    When I run the command "dk list <configFolder>"
    Then I should have a tab-delimited list with a datalink named "<name>" and fqdn "<fqdn>"
    And I should not have a tab-delimited list with a datalink named "<notName>" and fqdn "<notFqdn>"

  Scenario: list second data link
    Given the scenario "create second data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFolder | name       | fqdn                      | notFqdn                    | notName    |
      | secondOem    | datalink-2 | firstOem/datalink-2:1.0.1 | secondOem/datalink-1:1.0.0 | datalink-1 |
    When I run the command "dk list <configFolder>"
    Then I should have a tab-delimited list with a datalink named "<name>" and fqdn "<fqdn>"
    And I should not have a tab-delimited list with a datalink named "<notName>" and fqdn "<notFqdn>"
