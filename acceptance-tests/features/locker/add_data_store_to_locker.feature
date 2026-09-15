Feature: add data store to locker
  To be able to add a data store to a locker
  As an authenticated user
  I should be able to create a datalink, login to an account and publish the datalink
  I should be able to initialize a locker
  I should be able to add a data store to the locker for the account and datalink created

  Scenario: create data link for data store
    Given the following parameters
      | configFile                       | handle       | oem             | version |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-str-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

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
    And the scenario "create data link for data store" ran with condition "service_completed_successfully"
    And the following parameters
      | handle       | oem             | version |
      | locker-str-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: init data store locker
    Given the following parameters
      | path         | password |
      | myLocker.bin | SIzlR0a$ |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk init <path>"
    Then I should have a non-empty locker file initialized under "<path>"

  Scenario: add data store for incomplete data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the scenario "init data store locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password | handle      | dataLinkFqdn                 | dataLinkVersion | mtime |
      | myLocker.bin | SIzlR0a$ | myLockerStr | com.genaiz.test/locker-str-1 | 1.0.0           |       |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have an error "datalink property set is empty, is it incomplete?"

  Scenario: add data link secret property for data store
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem             | dataLinkVersion | key        | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-str-1 | com.genaiz.test | 1.0.0           | SECRET_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: re-publish data link
    Given the scenario "login data link" ran with condition "service_completed_successfully"
    And the scenario "create data link for data store" ran with condition "service_completed_successfully"
    And the following parameters
      | handle       | oem             | version | newVersion |
      | locker-str-1 | com.genaiz.test | 1.0.0   | 1.0.1      |
    When I run the command "dk publish <oem>/<handle>:<version> --new-version=<newVersion>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<newVersion>"

  Scenario: add data store for data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the scenario "init data store locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password | handle      | dataLinkFqdn                 | dataLinkVersion | mtime |
      | myLocker.bin | SIzlR0a$ | myLockerStr | com.genaiz.test/locker-str-1 | 1.0.1           |       |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"
