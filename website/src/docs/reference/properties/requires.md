# requires

Require specific Task variables to be set.

## Type

`requires_obj`


## Description

Declare required variables for the task. Task fails if variables are not provided.




## Contexts

This property can be used in:


- Task level








## Examples


### Require variables

Task needs specific variables

```yaml
tasks:
  deploy:
    requires:
      vars: [ENV, VERSION]
    cmds:
      - echo "Deploying {{.VERSION}} to {{.ENV}}"

```



### With error message

Custom error for missing variable

```yaml
tasks:
  release:
    requires:
      vars:
        - TAG
    cmds:
      - git tag {{.TAG}}

```






## Related

- [vars](./vars.md)
- [preconditions](./preconditions.md)




