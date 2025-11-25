Feature: data ports for the bash example
  To be able to add and remove data ports to the bash example function
  As a developer
  I should be able to create the bash example recipe, add an input port, add an output port and remove them both

  Scenario: create bash example
    Given the scenario "create bash example solution" ran with condition "service_completed_successfully"
    And the following parameters
      | recipe       | handle          | oem            | type     | version |
      | bash-example | my-bash-example | com.genaiz.dev | function | 1.1.1   |
    When I run the command "sf create <handle> --recipe=<recipe> --oem=<oem> --version=<version>"
    Then I should have a function under "<handle>" named "<handle>" with oem "<oem>" and version "<version>"
    And I should have a function under "<handle>" named "<handle>" with type "<type>"

  Scenario: add input data port by path
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | inputPath       | inputHandle | inputName | inputDescription |
      | my-bash-example | run/in/readPort | readPort    | Read Port | Read Port Test   |
    And the workdir changes to "<path>"
    When I run the command "sf data input add <inputPath> --name='<inputName>' --name='<inputDescription>'"
    Then I should have an input port under "<path>" with handle "<inputHandle>", named "<inputName>" with description "<inputDescription>"
    And I should have an empty folder under "<inputPath>"

  Scenario: add input data port by handle
    Given the scenario "create bash example" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | inputHandle |
      | my-bash-example | readPort2   |
    And the workdir changes to "<path>"
    When I run the command "sf data input add <inputHandle>"
    Then I should have an input port under "<path>" with handle "<inputHandle>", named "<inputHandle>" with description ""

  Scenario: remove input data port by handle
    Given the scenario "add input data port by handle" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | inputHandle |
      | my-bash-example | readPort2   |
    And the workdir changes to "<path>"
    When I run the command "sf data input rm <inputHandle>"
    Then I should not have an input port "<path>" with handle "<inputHandle>"

  Scenario: add output data port on an invalid path
    Given the scenario "remove input data port by handle" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | outputPath           | outputHandle |
      | my-bash-example | run/12/out/writePort | writePort    |
    And the workdir changes to "<path>"
    When I run the command "sf data output add <outputPath>"
    Then I should have an error for value "<outputPath>"

  Scenario add output data port by path
    Given the scenario "remove input data port by handle" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | outputPath           | outputHandle | outputName |
      | my-bash-example | run/12/out/writePort | writePort    | Write Port |
    And the workdir changes to "<path>"
    And the directory "<outputPath>" is created
    When I run the command "sf data output add <outputPath> --name='<outputName>'"

  Scenario: add output data port by handle
    Given the scenario "add output data port by path" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | outputHandle | outputDescription     |
      | my-bash-example | writePort2   | A description to keep |
    And the workdir changes to "<path>"
    When I run the command "sf data output add <outputHandle> --description=<outputDescription>"
    Then I should have an output port under "<path>" with handle "<outputHandle>", named "<outputHandle>" with description "<outputDescription>"

  Scenario: remove output data port by path
    Given the scenario "add output data port by handle" ran with condition "service_completed_successfully"
    And the following parameters
      | path            | outputPath            | outputHandle |
      | my-bash-example | run/12/out/writePort/ | writePort    |
    And the workdir changes to "<path>"
    When I run the command "sf data output rm <outputPath>"
    Then I should not have an output port "<path>" with handle "<outputHandle>"
