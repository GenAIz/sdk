Feature: locker init
  To be able to initialize a locker
  As a developer
  I should be able to init a new locker file, overwrite it, and update its credentials

  Scenario: init default user locker
    Given the following parameters
      | path                            | password   |
      | $HOME/.config/genaiz/locker.bin | myLocker9$ |
    When I run the command "lk init"
    And the parameter "<password>" entered on STDIN
    Then I should have a locker file under "<path>"

  Scenario: overwrite default user locker
    Given the scenario "init default user locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path                            | password   | mtime |
      | $HOME/.config/genaiz/locker.bin | myLocker9$ |       |
    And the modification time of "<path>" known as parameter "mtime"
    When I run the command "lk init --overwrite"
    And the parameter "<password>" entered on STDIN
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: update default user locker duplicate password
    Given the scenario "overwrite default user locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path                            | oldPassword | password   |
      | $HOME/.config/genaiz/locker.bin | myLocker9$  | myLocker9$ |
    When I run the command "lk init --update"
    And the parameter "<oldPassword>" entered on STDIN
    And the parameter "<password>" entered on STDIN
    Then I should have an error: "Error: the passphrase must be different"

  Scenario: update default user locker
    Given the scenario "overwrite default user locker" ran with condition "service_completed_successfully"
    And the following parameters
      | path                            | oldPassword | password     |
      | $HOME/.config/genaiz/locker.bin | myLocker9%  | yourLocKer7% |
    And the modification time of "<path>" known as parameter "mtime"
    When I run the command "lk init --update"
    And the parameter "<oldPassword>" entered on STDIN
    And the parameter "<password>" entered on STDIN
    Then I should have a locker file under "<path>" with a modification time different than "<mtime>"

  Scenario: init custom locker file
    Given the following parameters
      | path               | password   |
      | myCustomLocker.bin | myLocker9$ |
    When I run the command "lk init <path>"
    And the parameter "<password>" entered on STDIN
    Then I should have a locker file under "<path>"

  Scenario: init custom locker file, no overwrite, no update
    Given the following parameters
      | path               | overwrite | update |
      | myCustomLocker.bin | n         | n      |
    When I run the command "lk init <path>"
    And the parameter "<update>" entered on STDIN
    And the parameter "<overwrite>" entered on STDIN
    Then I should have an error: "Error the locker [<path>] is already initialized"
