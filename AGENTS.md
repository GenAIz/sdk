# AGENTS.md

## Validation policy

Validation is allowlist-only.

The agent may validate only the files, packages, and commands listed in this section.
Anything not explicitly listed is forbidden.

## Pull-request validation scope

When reviewing a pull request:

- Validate only files changed by the pull request relative to its target branch.
- Validate the commit message.
- Validate the pull request format.
- Do not validate unchanged files.
- Do not expand validation to imported packages, neighboring packages, or the entire repository.
- If no eligible files changed, do not run validation.
- If the changed-file list cannot be determined reliably, stop and report that validation was not run.
- When the validation is completed, if there are no rule violations, approve the pull request.

## Eligible Go files

- End in `.go`
- Do not end in `_test.go`

## Validation rule: commit message

- Report the commit message if it does not include an issue number from the repository.
- Report the commit message if it is written in a language that is not English

## Validation rule: pull-request format

- Report the pull request if it does not include an issue number from the repository in its title.
- Report the pull request if it is not linked with an issue under the Development section.

## Validation rule: logical inversions

- Report every condition which seems to be reversed and could lead to bugs.

## Validation rule: named returns

- Report every named return value.
- Report every naked return statements, unless used in void context.

## Validation rule: in-place slice mutations

- Report every usage of slices.DeleteFunc as bug-prone and discouraged, use slicez.Filter instead.
- Report every usage of in-place slice argument mutations as bug-prone and discouraged.

## Validation rule: hardcoded credentials

- Report every usage of potential credential strings hardcoded in Go source files.
