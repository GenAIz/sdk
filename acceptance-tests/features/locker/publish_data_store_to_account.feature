Feature: publish data store to account
  To be able to add a data store to a locker
  As an authenticated user
  I should be able to create a datalink, add prop specs, login to an account and publish the datalink
  I should be able to initialize a locker
  I should be able to add a data store to the locker for the account and datalink created
  I should be able to update the data store in the locker for the rop specs
  I should be able to publish the data store to the account owning the datalink

  Scenario: create data link for data store
    Given the following parameters
      | configFile                       | handle       | oem             | version |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-str-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link property
    Given the scenario "create data link for data store" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem             | version | key    | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-str-1 | com.genaiz.test | 1.0.0   | TENANT | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value ""

  Scenario: login data link for data store
    Given the orchestrator is running with condition: "service_healthy"
    And the environment contains "GENAIZ_PASSWORD=<password>"
    And the following parameters
      | password | path                     | username         |
      | success  | $HOME/.cache/genaiz.auth | _test@genaiz.com |
    When I run the command "ac login <orchestrator> --username=<username>"
    Then I should have an active session id with host "<orchestrator>" for username "<username>" under path "<path>"

  Scenario: publish data link
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the scenario "add data link secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | handle       | oem             | version |
      | locker-str-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: init data store locker
    Given the following parameters
      | path         | password   |
      | myLocker.bin | Testing37$ |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk init <path>"
    Then I should have a non-empty locker file initialized under "<path>"

  Scenario: add data store for data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the scenario "init data store locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      | dataLinkFqdn                 | dataLinkVersion | mtime |
      | myLocker.bin | Testing37$ | myLockerStr | com.genaiz.test/locker-str-1 | 1.0.0           |       |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data store property
    Given the scenario "add data store for data link" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      | key    | value    |
      | myLocker.bin | Testing37$ | myLockerStr | TENANT | myTenant |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str update <handle> <key> '<value>' --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: create data store for account
    Given the scenario "update data store property" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      |
      | myLocker.bin | Testing37$ | myLockerStr |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str publish <handle> --locker=<path>"
    Then I should have a data store named "<name>" created under account "<orchestrator>"

  Scenario: update data store property value
    Given the scenario "create data store for account" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      | key    | value            |
      | myLocker.bin | Testing37$ | myLockerStr | TENANT | some Store Value |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk src update <handle> <key> '<value>' --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data store for account
    Given the scenario "update data store property value" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      |
      | myLocker.bin | Testing37$ | myLockerStr |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk src publish <handle> --name='<name>' --locker=<path>"
    Then I should have a data store named "<name>" updated under account "<orchestrator>"
