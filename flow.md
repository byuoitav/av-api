+----------------------------Overview-----------------------------------+

+---------------+      +----------+       +------------+
|  Get commands | ---> | Order    | ----> |  Evaluate  |
|  for rooom    |      | commands |       |  Commands  |
|  configuration|      |          |       |            |
+---------------+      +----------+       +------------+
                                             |
       +------------------------------------ +
       |
       V
    +-----------+       +------------+
    | Reconcile | ----> |  Execute   |
    | Actions   |       |  actions   |
    +---------- +       +------------+

+-------------------------Evaluate Command------------------------------+
* An action is simply a struct in the Go source code that is passed
  between functions

This will require mapping
properties to commands.                                Performed by the
         |                                            /   "validate" function on
         |                                           /       the command struct.
         |                                          /
+---------------+      +-------------+       +------------+
|Check struct   | ---> |  Generate   | ----> |  validate  |
|for properties |      |  Actions    |       |  Action(s) |
|relating to cmd|      |             |       |            |
+---------------+      +-------------+       +------------+
            \          /                           |
             \        /                            |
              \      /                             |
               \    /                              |
         Performed by the               Need to know properties
        Evaluate function on              required of a device
       on the Command struct.             for a given action.
  (Includes generation of params)

+-------------------------Generate Actions-------------------------------+


+-------------------+                 +--------------------+
| Retrieve device   |                 | Get Action name    |
| based on building | --------------> |  based on property |
| room and name.    |                 |                    |
+-------------------+                 +--------------------+
                                              |
         +----------------------------------- +
         |
         V
+-------------------+
|   If needed add   |
|   property value  |
|   as parameter.   |
+-------------------+



+-------------------------Reconcile Actions------------------------------+
* Reconciling actions makes sure that commands are not executed in
  conflicting order

+------------------+       +------------------+      +---------------+
|  Build set of    |       | For each action  |      | Intersect sets|
| actions for each | ----> | get incompatable | ---> | if not empty, |
|      device      |       | commands.        |      | continue.     |
+------------------+       +------------------+      +---------------+
                                                              |
        +-----------If incompatable found---------------------+
        |
        V
+----------------------+       +------------------+
| Check if one of the  |       | Find which were  |
| incompatable actions | ----> | incompatable.    |
| are room-wide. If yes|       | Return error.    |
| override it.         |       |                  |
+----------------------+       +------------------+
