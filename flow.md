'''
+----------------------------Overview-----------------------------------+

+------------+      +----------+       +------------+
|  Get all   | ---> | Order    | ----> |  Evaluate  |
|  Commands  |      | commands |       |  Commands  |
+------------+      +----------+       +------------+
                                             |
       +------------------------------------ +
       |
       V
    +-----------+       +------------+
    | Reconcile | ----> |  Execute   |
    | Actions   |       |  actions   |
    +---------- +       +------------+

+-------------------------Evalutate Command------------------------------+

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
'''
