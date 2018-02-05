# Nomad-Shepard
A task restart coordinator for Hashicorp Nomad

## Purpose 
When utilizing the template stanza with change mode restart incombination with consul template various adverse restart conditions can occour in the cluster because they are uncoordinated. This can cause Thundering Herd Behavior as well as other types of unwanted ocillations to occour to running task groups and tasks with multiple instances.

### Approach Architecture
By watching the nomad allocation folder on the Nomad client machine instances it is possible to observe configuration changes in running tasks allocation folders. 

Using the allocation ID of the tasks allocation Nomad the allocation state is queired from the nomad client. When allocations achive the running state a watcher is set on the allocations folder and the allocations task group and task subfolders are watched. Currently shared_alloc folders and the local folders for each task will be tracked.

After tracking starts change events trigger shepard to take action to restart jobs. 
Each task can be made restarable by shepard by including the shepard.tpl file in the jobs local folder using the template stanza and setting the change method to restart.
In the shepard.tpl file include the following.

```
{keyOrDefault (printf "shepard/keys/%s/%s/%s/%s/%s" (env "NOMAD_NODE_NAME")(env "NOMAD_JOB_NAME")(env "NOMAD_GROUP_NAME")(env "NOMAD_TASK_NAME")(env "NOMAD_ALLOC_INDEX")) ""}}
```

By updating the value of this key a task can be restarted. As well as including this key and setting the shepard.tpl template to change mode restart all other templates populated through the template stanza should have their change modes set to no op.

Once shepard is watching a hosts allocation folders for a job or task any change events occour cause shepard to query alocation state through nomad and decide to aquire a lock on the following key value in consul along with setting its restart status to planning.
```
shepard/locks/${NOMAD_NODE_NAME}/${NOMAD_JOB_NAME}/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/task_lock
```

The shepard instance choosen to be leader will then choose how it allows the other shepard clients to restart their tasks.
This is done by creating multiple locks under task_locks namespace it will also update the number of concurrent locks present and also the number of task allocations aquired from the job configuration.

```
shepard/locks/${NOMAD_NODE_NAME}/${NOMAD_JOB_NAME}/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/restart_lock/1
```

Once the locks are commited to consul the task lock is updated with a restart status of planned. The leader will then release the lock and join the rest of the shepard and attempt to gain a restart_lock.

With the restart status set to planned each of the shepards aquires a restart lock and writes to the key value at
```
{keyOrDefault (printf "shepard/keys/%s/%s/%s/%s/%s" (env "NOMAD_NODE_NAME")(env "NOMAD_JOB_NAME")(env "NOMAD_GROUP_NAME")(env "NOMAD_TASK_NAME")(env "NOMAD_ALLOC_INDEX")) ""}}
```

This causes that allocations task to restart under nomad. Shepard then monitors the restart process until the task status is running and its health checks pass in consul. Then the task lock is aquired the node count is decremented and the task_lock and restart lock is released.

In addition to providing simple rolling restart capability the shepard client promoted to leader can choose how to enable the other task instances to restart based on the rolling update settings if defined in the nomad job spec.

## Running the application

- Install go lang for windows.
- Checkout the project into the users golang src folder
- cd to the project directory and run ```go get ./...```
- build the project with ```go install```
- update configuration in the config folder
- set SHEPARD_CONFIG_PATH
- ```shepard```

### Docker
- Coming soon









