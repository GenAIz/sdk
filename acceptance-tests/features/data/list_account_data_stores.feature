Feature: list account data stores
  To be able to list data stores available to an account
  As an authenticated user
  I should be able to login to a broker, create and publish a DataLink
  I should be able to initialize a locker and add a data store for the Datalink
  I should be able to publish the locker store and list it from the broker

  Scenario: create data link
    Given the following parameters
      | configFile                       | handle     | oem             | version |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <handle> --oem=<oem>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link property
    Given the scenario "create data link" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem             | version | key    | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   | TENANT | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value ""

  Scenario: add data link secret property
    Given the scenario "add data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle     | oem             | version | key        | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | datalink-1 | com.genaiz.test | 1.0.0   | SECRET_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: login account data stores
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
      | handle     | oem             | version |
      | datalink-1 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: init data store locker
    Given the following parameters
      | path         | password |
      | myLocker.bin | SIzlR0a$ |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk init <path>"
    Then I should have a non-empty locker file initialized under "<path>"

  Scenario: add data store for data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the scenario "init data store locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password | handle      | dataLinkFqdn               | dataLinkVersion | mtime |
      | myLocker.bin | SIzlR0a$ | myLockerStr | com.genaiz.test/datalink-1 | 1.0.0           |       |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data store property
    Given the scenario "add data str for data link" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | handle      | key    | value      |
      | myLocker.bin | myLockerStr | TENANT | some Value |
    And the modification time of "<path>" known as parameter "mtime"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str update <handle> <key> '<value>' --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data store secret property
    Given the scenario "update data store property" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password | localHandle | gpgPath  | gpgPassword | key        |
      | myLocker.bin | SIzlR0a$ | myLockerStr | test.gpg | gpgPass     | SECRET_KEY |
    And a gpg encrypted file copied on "<gpgPath>"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    # that's gonna need some special way of invoking the genaiz command from the docker image
    When I run the entrypoint "gpg --quiet --decrypt <gpgPath> 2>/dev/null | genaiz lk str update <localHandle> <key> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: create data store for account
    Given the scenario "update data store secret property" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password   | handle      | name         |
      | myLocker.bin | Testing37$ | myLockerStr | Locker Str 1 |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk str publish <handle> --name='<name>' --locker=<path>"
    Then I should have a data store named "<name>" created under account "<orchestrator>"

  Scenario: list data stores for account
    Given the scenario "create data store for account"
    And the following parameters
      | oem             | handle     | version | name         |
      | com.genaiz.test | datalink-1 | 1.0.0   | Locker Str 1 |
    When I run the command "dt str list <oem>/<handle>:<version>"
    Then I should have a data store list with first item listed named "<name>" for datalink "<oem>/<handle>:<version>" with "1" property
