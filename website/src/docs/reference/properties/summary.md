# summary

Long description shown with task --summary.

## Type

`string`


## Description

Detailed description of the task. Displayed when running `task --summary <task>`.




## Contexts

This property can be used in:


- Task level








## Examples


### Add summary

Provide detailed task documentation

```yaml
tasks:
  deploy:
    desc: Deploy application
    summary: |
      Deploys the application to production environment.
      
      Prerequisites:
      - Docker must be installed
      - AWS credentials configured
      
      This task will:
      1. Build the Docker image
      2. Push to ECR
      3. Update ECS service
    cmds:
      - task: build-image
      - task: push-image
      - task: update-service

```






## Related

- [desc](./desc.md)




