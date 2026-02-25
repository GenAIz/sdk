Feature: outbound proxy for the bash example
  To be able to add and remove outbound proxies to a bash example connector
  As an authenticated developer
  I should be able to create the bash example recipe, add and remove proxies from the connector and publish them

  Scenario: create bash function
    Given the following parameters
      | recipe       | handle           | oem            | type     | version |
      | bash-example | my-bash-function | com.genaiz.dev | function | 1.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --version=<version>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: add outbound proxy to function
    Given the scenario "create bash function" ran with condition "service_completed_successfully"
    And the following parameters
      | folder           | hostValue | hostPort |
      | my-bash-function | *         | 0        |
    And the workdir changes to "<folder>"
    When I run the command "sf proxy add <hostValue>:<hostPort>"
    Then I should have an error preventing outbound proxies from being configured

  Scenario: create bash connector
    Given the following parameters
      | recipe       | handle            | oem            | type      | version |
      | bash-example | my-bash-connector | com.genaiz.dev | connector | 1.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --version=<version> --type=<type>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>", version "<version>" and type "<type>"

  Scenario: add outbound proxy invalid host
    Given the scenario "create bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | hostValue    | hostPort |
      | my-bash-connector | $invalidHost | 22       |
    And the workdir changes to "<folder>"
    When I run the command "sf proxy add <hostValue>:<hostPort>"
    Then I should have an error for field "function.publish.proxyadd.host"

  Scenario: add outbound proxy invalid port
    Given the scenario "create bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | hostValue      | hostPort |
      | my-bash-connector | dev.genaiz.com | 0        |
      | my-bash-connector | dev.genaiz.com | 65536    |
    And the workdir changes to "<folder>"
    When I run the command "sf proxy add <hostValue>:<hostPort>"
    Then I should have an error for field "function.publish.proxyadd.port"

  Scenario: add outbound proxy
    Given the scenario "create bash connector" ran with condition "service_completed_successfully"
    And the following parameters
      | folder            | hostValue      | hostPort | options                | flags            |
      | my-bash-connector | dev.genaiz.com | 80       |                        | active, tcp      |
      | my-bash-connector | dev.genaiz.com | 22       | --tcp                  | active, tcp      |
      | my-bash-connector | dev.genaiz.com | 1        | --inactive             | tcp              |
      | my-bash-connector | dev.genaiz.com | 1024     | --udp                  | active, udp      |
      | my-bash-connector | dev.genaiz.com | 5074     | --inactive --udp       | udp              |
      | my-bash-connector | dev.genaiz.com | 8081     | --tcp --udp            | active, tcp, udp |
      | my-bash-connector | dev.genaiz.com | 65535    | --inactive --tcp --udp | tcp, udp         |
    And the workdir changes to "<folder>"
    When I run the command "sf proxy add <hostValue>:<hostPort> <options>"
    Then I should have an outbound proxy under "<folder>" with host "<hostValue>", port "<hostPort>" and flags "<flags>"
