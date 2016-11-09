1. Get room configuration for room from Database
  - Configuration contains a `Reconcile()` action
  - Configuration contains commands for the configuration
    - Each command contains
      - Priority
      - `Evaluate()` and `Validate()` action
1. Order commands in the configuration based on priority
1. For each command call:
  1. `Evaluate()`
    - `Evaluate()` takes a the information in the PUT body + room information and returns a list of actions corresponding to that command.
        - Can perform any sort of inner-command logic required for the configuration (i.e. scaling volume levels)
        - Each action contains
          - A device
          - A command name
          - A set of parameters (if necessary for the action)
  1. `Validate()`
        - `Validate()` takes the list of commands and checks if the commands/parameters are valid for the device specified. Returns true/false
        - If validate returns true, add to compiled list of actions from all commands
1. Once list of actions has been compiled for all commands in the configuration, call `Reconcile()`
    - `Reconcile()` takes the list of actions and can perform any sort of checking, reordering, or any other configuration specific actions that are inter-action, and returns an ordered list of actions. (the logic is dependent on the interplay of actions/commands. i.e. Always turn D1 on before D2)
1. For each action in the list, execute it.
1. Return new state + report of actions taken.

## Room Configurations

The AV API uses logic surrounding the interaction with devices to make determinations of command execution  refer to rules governing different room types as **room configurations**.

A room configuration



### Mapping state to commands

The AV API is a RESTful service. Interaction occurs by requesting and setting the state of resources. Sending a PUT request with the JSON payload of:

```
{
  "displays": [{
      "name": "D1",
      "power": "on"
    }]  
}
```

To the endpoint `/buildings/ITB/rooms/1001D` is saying "in room 1001D in the ITB, set the power state of display D1 to on."

It is the purpose of the AV API to generate, and then execute the **command** (or commands) required to place the room into the state requested by the user. In the AV API commands are executed by calling RPC endpoints on microservices that correspond to a command and protocol.

Thus the body

```
{
  "displays": [{
      "name": "D1",
      "power": "on"
    }]  
}
```

Must be translated to the command:

```
  http://sony-control-microservice/10.10.10.10/power/on
```

Implying the intermediary mappings of

1. Device `D1` in room `ITB-1001D` is a device that can be communicated with via the `sony-control-microservice`.
1. Device `D1` has address of `10.10.10.10`
1. The property `power` with the value of `on` corresponds to the command endpoint `power/on`

The AV-API maintains these mappings, as well as any logic surrounding command execution for a given room.



#### Logic surrounding multiple command execution

Let's say that there exists a room `ITB-1001D` containing a London-DSP audio device, 2 Sony TV's, and a projector. For reasons specific to this room

For example the PUT body

```
{
    "power": "on"
}  
```

Sent to the endpoint `/buildings/ITB/rooms/1001D` requests that the entire room's `power` state be set to `on`. In ITB-1001D the room-wide power property aliases to each individual device's power state (for more on customizing command-mapping within a room see <<ROOM CONFIGURATION LINK>>). Thus

```
{
    "power": "on"
}  
```

Must be translated to the set of commands:

```
[
  http://sony-control-microservice/10.10.10.3/power/on,
  http://pjlinkmicroservice/10.10.10.2/power/on,
  http://sony-control-microservice/10.10.10.5/power/on,
  http://londondsp-control-microservice/10.10.10.22/power/on
]
```

The AV-API maintains the logic required to set
