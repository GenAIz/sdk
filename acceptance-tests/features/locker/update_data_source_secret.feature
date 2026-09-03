Feature: update data source secret to locker
  To be able to update a data source secret to a locker
  As a developer
  I should be able to create a datalink, add regular and secret prop specs, login to an account and publish the datalink
  I should be able to initialize a locker
  I should be able to add a data source to the locker for the account and datalink created
  I should be able to update the data source in the locker for the regular and secret prop specs

  Scenario: create data link for data source
    Given the following parameters
      | configFile                       | handle       | oem             | version |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-src-2 | com.genaiz.test | 1.0.0   |
    And the user genaiz config folder is under <path>
    When I run the command "dk create <oem>/<handle>"
    Then I should have a datalink under "<configFile>" named "<handle>", with handle "<handle>", oem "<oem>" and version "<version>"

  Scenario: add data link property
    Given the scenario "create data link for data source" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem             | version | key      | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-src-2 | com.genaiz.test | 1.0.0   | TEST_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key>"
    Then I should have a "<type>" property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>" and default value ""

  Scenario: add data link secret property
    Given the scenario "add data link property" ran with condition "service_completed_successfully"
    And the following parameters
      | configFile                       | handle       | oem             | version | key        | type   |
      | $HOME/.config/genaiz/Genaiz.yaml | locker-src-2 | com.genaiz.test | 1.0.0   | SECRET_KEY | STRING |
    When I run the command "dk prop add <oem>/<handle>:<version> <key> --secret"
    Then I should have a "<type>" secret property spec under "<configFile>", for a datalink with handle "<handle>", oem "<oem>" and version "<version>", with key "<key>"

  Scenario: login data link for data source
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
      | locker-src-2 | com.genaiz.test | 1.0.0   |
    When I run the command "dk publish <oem>/<handle>:<version>"
    Then I should have a datalink published to the orchestrator with fqdn "<oem>/<handle>:<version>"

  Scenario: init data source locker
    Given the following parameters
      | path         | password |
      | myLocker.bin | SIzlR0a$ |
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    When I run the command "lk init <path>"
    Then I should have a non-empty locker file initialized under "<path>"

  Scenario: add data source for data link
    Given the scenario "publish data link" ran with condition "service_completed_successfully"
    And the scenario "init data source locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | handle      | dataLinkFqdn                 | dataLinkVersion | mtime |
      | myLocker.bin | myLockerSrc | com.genaiz.test/locker-src-2 | 1.0.0           |       |
    And the modification time of "<path>" known as parameter "mtime"
    When I run the command "lk src add <handle> <dataLinkFqdn>:<dataLinkVersion> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data source property
    Given the scenario "add data source for data link" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | handle      | key      | value      |
      | myLocker.bin | myLockerSrc | TEST_KEY | some Value |
    And the modification time of "<path>" known as parameter "mtime"
    When I run the command "lk src update <handle> <key> '<value>' --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update data source secret property
    Given the scenario "update data source property" ran with condition "service_completed_successfully"
    And the following parameters
      | path         | password | localHandle | gpgPath  | gpgPassword | key           |
      | myLocker.bin | SIzlR0a$ | myLockerSrc | test.gpg | gpgPass     | myLocalHandle |
    And a gpg encrypted file copied on "<gpgPath>"
    And the environment contains "GENAIZ_LK_PASSWORD=<password>"
    # that's gonna need some special way of invoking the genaiz command from the docker image
    When I run the entrypoint "gpg --quiet --decrypt <gpgPath> 2>/dev/null | genaiz lk src update <localHandle> <key> --locker=<path>"
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"
