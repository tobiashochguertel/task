# shopt

Enable bash shell options.

## Type

`array`


## Description

Set bash-specific shell options using shopt. Alternative to set for bash users.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level








## Examples


### Enable globstar

Use ** for recursive globbing

```yaml
tasks:
  clean:
    shopt: [globstar]
    cmds:
      - rm -rf **/*.pyc

```



### Multiple options

Enable several bash options

```yaml
version: '3'

shopt: [globstar, dotglob]

tasks:
  lint:
    cmds:
      - shellcheck **/*.sh

```






## Related

- [set](./set.md)




