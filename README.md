# Nomad-Shepard
A task restart coordinator for Hashicorp Nomad

## Project Status
 - Allocation watcher: In Progress
 - Restart Planner: To Do
 - Restart Coordinator: To Do
 - Garbage Collection: To Do
 - Log Forwarding: To Do

## Purpose 
When utilizing the template stanza with change mode restart incombination with consul template various adverse restart conditions can occour in the cluster because they are uncoordinated. This can cause Thundering Herd Behavior as well as other types of unwanted ocillations to occour to running task groups and tasks with multiple instances.

### Functionality
#### Restart Coordination
By watching the nomad allocation folder on the Nomad client machine instances it is possible to observe configuration changes in running tasks allocation folders. 

Using the allocation ID of the tasks allocation state is aquired from the Nomad client. When allocations achive the running state and the task includes the shepard.tpl file in its local folder a watcher is set on the allocations folder. Currently a task groups shared_alloc folder and the local folders for each task will be tracked. Once tracking starts, change events trigger a Shepard instance running on the client to take action to plan a restart and restart the job. 

Each task is made restartable by Shepard by including the shepard.tpl file in the jobs local folder with Nomads template stanza and setting the change method to restart. In the shepard.tpl file include the following Consul template statement.
```
{keyOrDefault (printf "shepard/keys/%s/%s/%s/%s/%s" (env "NOMAD_NODE_NAME")(env "NOMAD_JOB_NAME")(env "NOMAD_GROUP_NAME")(env "NOMAD_TASK_NAME")(env "NOMAD_ALLOC_INDEX")) ""}}
```
By updating the value of this key in Consul a task can be restarted. Additionally all other templates populated through Nomads template stanza should have their change modes set to no-op so that Shepard fully controls how a task restarts from configuration changes.

Once shepard is watching a hosts allocation folders for a job or task any change events that occour cause shepard to query allocation state through nomad and decide to aquire a lock on the following key value in consul. 
```
shepard/locks/${NOMAD_NODE_NAME}/${NOMAD_JOB_NAME}/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/task_lock
```
The Shepard instance choosen to be leader will then create the plan other shepard clients restarting this jobs task will follow.
The aquired lock is updated with the tasks instance_count, concurrent_lock_count, status=planned. The leader then releases the lock and joins the rest of the Shepards in attempting to aquire a restart_lock at. Other Shepard instances poll the task_lock for status while the leader reads job configuration from nomad and sets the restart plan.



With the restart status set to planned each of the Shepards attempts to aquire a restart lock by iterating through the count of concurrent locks and trying to aquire it.
```
shepard/locks/${NOMAD_NODE_NAME}/${NOMAD_JOB_NAME}/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/restart_lock/{{lock number}}
```

If a restart lock is aquired the Shepard instance writes a random value to consul at the following path.
```
shepard/keys/${NOMAD_NODE_NAME}/${NOMAD_JOB_NAME}/${NOMAD_GROUP_NAME}/${NOMAD_TASK_NAME}/${NOMAD_ALLOC_INDEX}"
```
This causes that allocations task to restart under nomad. Shepard then monitors the restart process until the task status is running and its health checks pass in consul. Then the task lock is aquired the node count is decremented and the task_lock and restart lock is released.

#### Garbage collection
Shepard instance will perodically poll for job and allocation keys in consul and check their status in Nomad. If any keys are present for allocations or jobs that are no longer in the running state on that node they will be removed from Consul.

In addition to job and allocation cleanup for a node in Consuls key store Shepard will cleanup its keystore values for nodes that no longer exist or have been dead for longer than the ```node_cleanup_timeout``` key in configuration file. 

#### Log shipping
Because Shepard instances have access to the allocation folders of a task it is possible to tag log files with meta tags and pass them to a log shipper for log aggragation based on a schedual. The schedual can be set to ship logs from a client every duration interval by setting the ```ship_log_interval``` key in the configuration file 


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









