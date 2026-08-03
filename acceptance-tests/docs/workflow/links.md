## Workflow Links

Links is a command used to manage workflow links between [nodes](nodes.md) on solutions to be published. A solution can
be viewed as a set of workflows each dictating one or more series of functions or repositories to be invoked in a
certain sequence.

With the current design of the SDK a workflow link can only be established between 2 existing nodes of the same
solution. There is always a left participating node and a right participant.

> [!IMPORTANT]
> The current version of the SDK does not validate whether data ports specified with the links exist on the referred
> functions if any.

### link strings

Both [links add](#links-add) and [links remove](#links-remove-rm) accept a list of strings describing a serialized link:

```
LEFT[[PORT]]:RIGHT[[PORT]]
```

* When **LEFT** and/or **RIGHT** refers to node handles within the Genaiz.yaml solution file
    * The **[PORT]** strings then refer to data ports supported by these nodes
* When **LEFT** and/or **RIGHT** refer to a relative Smart Function **PATH**
    * if the path contains a leaf folder, it is interpreted as the **[PORT]** suffix ex: myFunction/run/DataPort, the
      port will be resolved to **[DataPort]**
    * if the path is a single folder resolving to a function, the **[PORT]** suffix will still be read and eventually
      validated
    * if the string contains both a leaf folder and a **[PORT]** suffix, the commands will return an error, unless they
      are the same. (Resolution conflict)
* A **[PORT]** suffix or a leaf folder are not necessary to resolve or establish a link

#### Examples

```
genaiz wf links add myWorkflow myFunction1/run/out/outputPort:myFunction2
genaiz wf links add myWorkflow myFunction1[outputPort]:myNode3[nodePort]
genaiz wf links add myWorkflow myNode[nodePort]:myNode2[nodePort]
```

These should be invalid:

```
genaiz wf links add myWorkflow myFunction1/run/out/outputPort[outputPort2]:myNode2
genaiz wf links add myWorkflow myNode1[/run/out/outputPort]:myFunction2/run/in/inputPort
```

### links add

```
genaiz wf links add WORKFLOW_HANDLE LEFT[[PORT]]:RIGHT[[PORT]] \
  [LEFT[[PORT]]...] --config-type=TYPE
```

The add command allows adding from 1 to n links with 2 forms of specification: The direct node handle names and a path
string where both the handle and port can be parsed from a Smart Function configuration.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error:
  `Error: workflow hande [...] not found`

#### LEFT, RIGHT

* if either LEFT or RIGHT does not resolve to a node within the Solution or a Smart Function folder, the command returns
  an error: `Error: value [...] could not resolve to a workflow node`

#### [PORT]

* if LEFT and its [PORT] specification resolved two different port string, the command returns an error:
  `Error: conflicting port specification: [...] and [...] diverge`

### links remove (rm)

```
genaiz wf links rm WORKFLOW_HANDLE LEFT[[PORT]]:RIGHT[[PORT]] \
  [LEFT[[PORT]]...] --config-type=TYPE
```

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error:
  `Error: workflow hande [...] not found`

#### LEFT, RIGHT

* if either LEFT or RIGHT does not resolve to a node within the Solution or a Smart Function folder, the command returns
  an error: `Error: value [...] could not resolve to a workflow node`

#### [PORT]

* if LEFT and its [PORT] specification resolved two different port string, the command returns an error:
  `Error: conflicting port specification: [...] and [...] diverge`
